# darwinx

A simple Postgres schema evolution library for Go, using the [pgx](https://github.com/jackc/pgx) driver. Inspired by
the [darwin](https://github.com/GuiaBolso/darwin) library from GuiaBolso.

## Features

- Easy schema migration
- Postgres-specific
- Designed for simple Go applications
- Supports upgrade-only migrations (no branching)

Migration versions are stored as `DOUBLE PRECISION` / `float64` in Go. Prefer integer-style values (for example `1`, `2`, `3`) to avoid floating-point surprises.

The optional table name (`WithTableName`) must be a single unquoted SQL identifier (letters, digits, underscore only); it is not intended for schema-qualified names such as `public.migration`.

`Info` creates the migration history table if needed (same as `Migrate`), then returns status for each configured migration.

## Installation

To install DarwinX, use `go get`:

```bash
go get github.com/Doomsta/darwinx
```

## Usage

Here is a basic example:

```go
package main

import (
	"context"
	"github.com/Doomsta/darwinx"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	migrations = []darwinx.Migration{{
		Version:     1,
		Description: "Creating table posts",
		Script: `CREATE TABLE posts (
            id INT,
            title VARCHAR(255),
            PRIMARY KEY (id)
        );`,
	}}
)

func main() {
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, /* postgresURL */)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	d, err := darwinx.New(conn, darwinx.WithMigration(migrations))
	if err != nil {
		// Handle error
	}
	err = d.Migrate(ctx)
	if err != nil {
		// Handle error
	}
}
```

For more advanced usage and working examples, see the `example/` directory.

## Core Functionality

The library handles the following processes:

1. **Create a Table**: If it doesn't exist, a migration history table is created to track all applied migrations:
   ```sql
   CREATE TABLE IF NOT EXISTS %s (
       id             INT GENERATED ALWAYS AS IDENTITY NOT NULL,
       version        DOUBLE PRECISION NOT NULL,
       description    TEXT NOT NULL,
       checksum       CHARACTER(64) NOT NULL CHECK (checksum <> ''),
       applied_at     TIMESTAMPTZ NOT NULL,
       execution_time INTERVAL NOT NULL,
       UNIQUE(version),
       PRIMARY KEY (id)
   );
   ```

2. **Collect and Sort Migrations**: The provided migrations are collected and sorted by version.

3. **Compare with Database**: The collected migrations are compared against the current state of the database.

4. **Apply Migrations**: New migrations are applied in order, and the migration history table is updated accordingly.

## Testing

Unit tests:

```bash
go test . -short
```

Integration tests (Postgres via [Testcontainers](https://golang.testcontainers.org/), requires Docker) omit `-short`:

```bash
go test .
```

If Docker is unavailable, integration tests are skipped automatically.

## Contributing

Contributions are welcome! Please fork the repository and use a feature branch. Pull requests are warmly welcome.

## License

This project is licensed under the MIT License.