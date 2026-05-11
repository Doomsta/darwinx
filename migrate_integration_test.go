package darwinx

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func startPostgres(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, func()) {
	t.Helper()
	if testing.Short() {
		t.Skip("Docker-backed integration test (omit -short to run)")
	}
	pg, err := postgres.Run(ctx, "postgres:18-alpine")
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "docker") || strings.Contains(msg, "provider") {
			t.Skipf("Docker/Testcontainers not available: %v", err)
		}
		require.NoError(t, err)
	}
	cleanup := func() { _ = pg.Terminate(ctx) }
	return pg, cleanup
}

func TestIntegration_MigrateHappyPath(t *testing.T) {
	ctx := context.Background()
	pg, cleanup := startPostgres(t, ctx)
	t.Cleanup(cleanup)

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	migrations := []Migration{
		{Version: 1, Description: "create t", Script: `CREATE TABLE t (id int PRIMARY KEY);`},
		{Version: 2, Description: "alter t", Script: `ALTER TABLE t ADD COLUMN n text;`},
	}
	d, err := New(pool, WithMigration(migrations))
	require.NoError(t, err)
	require.NoError(t, d.Migrate(ctx))
	require.NoError(t, d.Migrate(ctx))
	require.NoError(t, d.Validate(ctx))

	records, err := d.allFromDB(ctx)
	require.NoError(t, err)
	require.Len(t, records, 2)

	info, err := d.Info(ctx)
	require.NoError(t, err)
	require.Len(t, info, 2)
	for _, row := range info {
		require.Equal(t, Applied, row.Status)
	}
}

func TestIntegration_MigrateRollbackSingleTransaction(t *testing.T) {
	ctx := context.Background()
	pg, cleanup := startPostgres(t, ctx)
	t.Cleanup(cleanup)

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	migrations := []Migration{
		{Version: 1, Description: "ok", Script: `CREATE TABLE t (id int PRIMARY KEY);`},
		{Version: 2, Description: "bad sql", Script: `THIS IS NOT VALID SQL;`},
	}
	d, err := New(pool, WithMigration(migrations))
	require.NoError(t, err)
	require.Error(t, d.Migrate(ctx))

	var cnt int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM migration`).Scan(&cnt))
	require.Equal(t, 0, cnt)

	var relExists bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_class WHERE relname = 't')`,
	).Scan(&relExists))
	require.False(t, relExists, "CREATE TABLE from first migration should be rolled back")
}

func TestIntegration_ValidateChecksumMismatch(t *testing.T) {
	ctx := context.Background()
	pg, cleanup := startPostgres(t, ctx)
	t.Cleanup(cleanup)

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	migrations := []Migration{
		{Version: 1, Description: "create t", Script: `CREATE TABLE t (id int PRIMARY KEY);`},
	}
	d, err := New(pool, WithMigration(migrations))
	require.NoError(t, err)
	require.NoError(t, d.Migrate(ctx))

	d2, err := New(pool, WithMigration([]Migration{
		{Version: 1, Description: "create t", Script: `CREATE TABLE t (id bigint PRIMARY KEY);`},
	}))
	require.NoError(t, err)
	err = d2.Validate(ctx)
	require.Error(t, err)
	var inv InvalidChecksumError
	require.True(t, errors.As(err, &inv), "expected InvalidChecksumError, got %T: %v", err, err)
	require.Contains(t, inv.Versions, float64(1))
}

func TestIntegration_MigrateWithoutTransactionPartialApply(t *testing.T) {
	ctx := context.Background()
	pg, cleanup := startPostgres(t, ctx)
	t.Cleanup(cleanup)

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	migrations := []Migration{
		{Version: 1, Description: "ok", Script: `CREATE TABLE t (id int PRIMARY KEY);`},
		{Version: 2, Description: "bad sql", Script: `THIS IS NOT VALID SQL;`},
	}
	d, err := New(pool, WithMigration(migrations), WithNoTransaction())
	require.NoError(t, err)
	require.Error(t, d.Migrate(ctx))

	var cnt int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM migration`).Scan(&cnt))
	require.Equal(t, 1, cnt, "first migration commits before second fails")

	var relExists bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_class WHERE relname = 't')`,
	).Scan(&relExists))
	require.True(t, relExists)
}
