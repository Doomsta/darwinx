# darwinx

Postgres schema evolution library for Go with [pgx](https://github.com/jackc/pgx)  
Based on https://github.com/GuiaBolso/darwin
# Example

```go
package main

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"net/url"
)

var (
	migrations = []darwinx.Migration{
		{
			Version:     1,
			Description: "Creating table posts",
			Script: `CREATE TABLE posts (
						id INT 		auto_increment, 
						title 		VARCHAR(255),
						PRIMARY KEY (id)
					 );`,
		},
		{
			Version:     2,
			Description: "Adding column body",
			Script:      "ALTER TABLE posts ADD body TEXT AFTER title;",
		},
	}
)

func main() {
	conn, _ := pgx.Connect(context.Background(), (&url.URL{
        Scheme:   "postgres",
        User:     url.UserPassword("postgres", "admin"),
        Host:     "localhost:5432/tcp",
        Path:     "db",
        RawQuery: "sslmode=disable&timezone=UTC",
    }).String())

	d, _ := darwinx.New(conn, darwinx.WithMigration(migrations))

	err := d.Migrate()
	if err != nil {
		log.Println(err)
	}
}
```
