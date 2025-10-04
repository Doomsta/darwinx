package integration

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/Doomsta/darwinx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed migrations.sql
var migrations []byte

type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}

func Test_Integration(t *testing.T) {
	ctx := context.Background()
	pgC, err := CreatePostgresContainer(ctx)
	if err != nil {
		require.NoError(t, err, "failed to create postgres container")
	}
	defer func() {
		if err := pgC.Terminate(ctx); err != nil {
			require.NoError(t, err, "failed to terminate container")
		}
	}()

	conn, err := pgxpool.New(ctx, pgC.ConnectionString)
	require.NoError(t, err)

	m, err := darwinx.MigrationsFromString(string(migrations))
	require.NoError(t, err)

	d, err := darwinx.New(conn, darwinx.WithMigration(m))
	require.NoError(t, err)

	err = d.Migrate(ctx)
	require.NoError(t, err)

	infos, err := d.Records(ctx)
	require.NoError(t, err)

	for _, info := range infos {
		fmt.Println("-", info.Version, info.Description, info.AppliedAt, info.ExecutionTime)
	}
}
