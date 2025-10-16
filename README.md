# darwinx

A simple Postgres schema evolution library for Go, using the [pgx](https://github.com/jackc/pgx) driver. Inspired by
the [darwin](https://github.com/GuiaBolso/darwin) library from GuiaBolso.

## Features

- Easy schema migration 
- Postgres-specific
- Designed for simple Go applications
supports upgrade-only, no branching migrations

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
	conn, _ := pgxpool.New(ctx, /* postgresURL */)

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

## Contributing

Contributions are welcome! Please fork the repository and use a feature branch. Pull requests are warmly welcome.

## License

This project is licensed under the MIT License.