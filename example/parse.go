package main

import (
	"context"
	"github.com/Doomsta/darwinx"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
)

const txt = `---- 1 Creating table posts
CREATE TABLE posts (
	id INT,
	title VARCHAR(255),
	PRIMARY KEY (id)
);

---- 2 Adding column body
ALTER TABLE posts ADD body TEXT AFTER title;`

func main() {
	ctx := context.Background()
	conn, _ := pgxpool.New(ctx, (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("postgres", "admin"),
		Host:     "localhost:5432/tcp",
		Path:     "db",
		RawQuery: "sslmode=disable&timezone=UTC",
	}).String())

	migrations, _ := darwinx.MigrationsFromString(txt)

	d, _ := darwinx.New(conn, darwinx.WithMigration(migrations))

	_ = d.Migrate(ctx)
}
