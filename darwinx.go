package darwinx

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"sort"
	"sync"
	"time"
)

const selectQuery = `SELECT
	version,
	description,
	checksum,
	applied_at,
	execution_time
FROM %s
	ORDER BY version ASC;`

const schemaQuery = `CREATE TABLE IF NOT EXISTS %s (
	id             INT GENERATED ALWAYS AS IDENTITY NOT NULL,
	version        DOUBLE PRECISION NOT NULL,
	description    TEXT NOT NULL,
	checksum       CHARACTER(128) NOT NULL CHECK (checksum <> ''),
	applied_at     TIMESTAMPTZ NOT NULL,
	execution_time INTERVAL NOT NULL,
	UNIQUE(version),
	PRIMARY KEY (id)
);`

const insertQuery = `INSERT INTO %s (
	version,
	description,
	checksum,
	applied_at,
	execution_time
) VALUES ($1, $2, $3, $4, $5);`

// Darwinx is a helper struct to access the Validate and migration functions
type Darwinx struct {
	conn       *pgx.Conn
	migrations []Migration
	mutex      sync.Mutex
	tableName  string
}

// New returns a new Darwinx struct
func New(conn *pgx.Conn, options ...Option) (*Darwinx, error) {
	d := &Darwinx{
		conn:      conn,
		mutex:     sync.Mutex{},
		tableName: "migration",
	}
	for _, o := range options {
		err := o.apply(d)
		if err != nil {
			return nil, err
		}
	}
	return d, nil
}

// Migrate executes the missing migrations in database
// Apply all Migration or rollback
func (d *Darwinx) Migrate(ctx context.Context) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	err := d.createTable(ctx)
	if err != nil {
		return err
	}

	err = d.Validate(ctx)
	if err != nil {
		return err
	}

	records, err := d.allFromDB(ctx)
	if err != nil {
		return err
	}

	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, migration := range planMigration(records, d.migrations) {
		s := time.Now()
		_, err := tx.Exec(ctx, migration.Script)
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				panic(err)
			}
			return errors.WithStack(err)
		}
		err = d.insertRecord(ctx, tx, MigrationRecord{
			Version:       migration.Version,
			Description:   migration.Description,
			Checksum:      migration.Checksum(),
			AppliedAt:     s,
			ExecutionTime: time.Since(s),
		})
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				panic(err)
			}
			return err
		}
	}

	return tx.Commit(ctx)
}

// Info returns the status of all migrations
func (d *Darwinx) Info(ctx context.Context) ([]MigrationInfo, error) {
	var info []MigrationInfo
	records, err := d.allFromDB(ctx)
	if err != nil {
		return info, err
	}
	sort.Sort(sort.Reverse(byMigrationRecordVersion(records)))

	for _, migration := range d.migrations {
		info = append(info, MigrationInfo{
			Status:    getStatus(records, migration),
			Error:     nil,
			Migration: migration,
		})
	}
	return info, nil
}

func (d *Darwinx) allFromDB(ctx context.Context) ([]MigrationRecord, error) {
	query := fmt.Sprintf(selectQuery, d.tableName)
	rows, err := d.conn.Query(ctx, query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	var migrations []MigrationRecord
	for rows.Next() {
		record := MigrationRecord{}
		err := rows.Scan(&record.Version, &record.Description, &record.Checksum, &record.AppliedAt, &record.ExecutionTime)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		migrations = append(migrations, record)
	}
	return migrations, nil
}

func (d *Darwinx) createTable(ctx context.Context) error {
	_, err := d.conn.Exec(ctx, fmt.Sprintf(schemaQuery, d.tableName))
	return errors.WithStack(err)
}

func (d *Darwinx) insertRecord(ctx context.Context, tx pgx.Tx, record MigrationRecord) error {
	_, err := tx.Exec(ctx, fmt.Sprintf(insertQuery, d.tableName),
		record.Version,
		record.Description,
		record.Checksum,
		record.AppliedAt,
		record.ExecutionTime,
	)
	return err
}

func planMigration(records []MigrationRecord, migrations []Migration) []Migration {
	if len(records) == 0 {
		return migrations
	}

	var planned []Migration

	last := -.1
	for _, r := range records {
		if r.Version > last {
			last = r.Version
		}
	}

	// Apply all migrations that are greater than the last migration
	for _, migration := range migrations {
		if migration.Version > last {
			planned = append(planned, migration)
		}
	}

	sort.Sort(byMigrationVersion(planned))

	return planned
}
