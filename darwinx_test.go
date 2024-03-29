package darwinx

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestMigrate(t *testing.T) {
	ctx := context.Background()
	migrations := []Migration{
		{
			Version:     1,
			Description: "Version 1",
			Script:      "CREATE TABLE foo (id SERIAL)",
		},
	}
	conn, teardown, err := setup(ctx)
	defer teardown()
	assert.NoError(t, err)
	assert.NoError(t, conn.Ping(context.Background()))
	d, err := New(conn, WithMigration(migrations))
	assert.NoError(t, err)
	err = d.Migrate(context.Background())
	assert.NoError(t, err)

	info, err := d.Info(context.Background())
	assert.NoError(t, err)
	assert.Len(t, info, 1)
	assert.NoError(t, info[0].Error)
	assert.Equal(t, info[0].Status, Applied)
}

func TestMigrate2Runs(t *testing.T) {
	ctx := context.Background()

	migrations := []Migration{
		{
			Version:     1,
			Description: "Version 1",
			Script:      "CREATE TABLE foo (id SERIAL)",
		},
		{
			Version:     2,
			Description: "Version 2",
			Script:      "CREATE TABLE foo2 (id SERIAL)",
		},
	}
	conn, teardown, err := setup(ctx)
	defer teardown()
	assert.NoError(t, err)
	assert.NoError(t, conn.Ping(ctx))
	d, err := New(conn, WithMigration(migrations))
	assert.NoError(t, err)
	err = d.Migrate(ctx)
	assert.NoError(t, err)
	err = d.Migrate(ctx)
	assert.NoError(t, err)

	_, err = d.Info(ctx)
	assert.NoError(t, err)
}

func TestMigrateInvalidSQL(t *testing.T) {
	ctx := context.Background()

	migrations := []Migration{
		{
			Version:     1,
			Description: "Version 1",
			Script:      "CREATE TABLE foo (id SERIAL)",
		},
		{
			Version:     2,
			Description: "Version 2",
			Script:      "CREATE haehfj // TABLE --foo (id SERIAL)",
		},
	}
	conn, teardown, err := setup(ctx)
	defer teardown()
	assert.NoError(t, err)
	assert.NoError(t, conn.Ping(ctx))
	d, err := New(conn, WithMigration(migrations))
	assert.NoError(t, err)
	err = d.Migrate(ctx)
	assert.Error(t, err)
}

func setup(ctx context.Context) (*pgx.Conn, func() error, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, func() error {
			return nil
		}, errors.Errorf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=admin", "POSTGRES_DB=test"})
	if err != nil {
		return nil, func() error {
			return nil
		}, errors.Errorf("Could not start resource: %s", err)
	}
	_ = resource.Expire(60) // Tell docker to hard kill the container in 60 seconds
	var conn *pgx.Conn
	if err := pool.Retry(func() error {
		var err error
		conn, err = pgx.Connect(ctx, (&url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword("postgres", "admin"),
			Host:     fmt.Sprintf("localhost:%s", resource.GetPort("5432/tcp")),
			Path:     "test",
			RawQuery: "sslmode=disable&timezone=UTC",
		}).String())
		if err != nil {
			return errors.WithStack(err)
		}
		return conn.Ping(ctx)
	}); err != nil {
		return nil, func() error {
			return nil
		}, errors.Errorf("Could not connect to docker: %s", err)
	}
	return conn, func() error {
		if err := pool.Purge(resource); err != nil {
			return errors.Errorf("Could not purge resource: %s", err)
		}
		return nil
	}, nil
}
