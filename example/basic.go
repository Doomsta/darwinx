package main

import (
	"context"
	"github.com/Doomsta/darwinx"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
)

var (
	migrations = []darwinx.Migration{{
		Version:     1,
		Description: "Creating table posts",
		Script: `CREATE TABLE posts (
    id INT 		auto_increment, 
	title 		VARCHAR(255),
	PRIMARY KEY (id)
 );`,
	}, {
		Version:     2,
		Description: "Adding column body",
		Script:      "ALTER TABLE posts ADD body TEXT AFTER title;",
	}}
)

func main() {
	ctx := context.Background()
	conn, _ := pgxpool.New(ctx, (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("postgres", "admin"),
		Host:     "localhost:5432/tcp",
		Path:     "db",
		RawQuery: "sslmode=disable&timezone=UTC",
	}).String())

	d, _ := darwinx.New(conn, darwinx.WithMigration(migrations))

	_ = d.Migrate(ctx)
}
