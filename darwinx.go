package darwinx

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"sort"
	"sync"
	"time"
)

const schemaQuery = `CREATE TABLE IF NOT EXISTS %s (
	id             INT GENERATED ALWAYS AS IDENTITY NOT NULL,
	version        DOUBLE PRECISION NOT NULL,
	description    TEXT NOT NULL,
	checksum       CHARACTER(64) NOT NULL CHECK (checksum <> ''),
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
	pool       *pgxpool.Pool
	migrations []Migration
	mutex      sync.Mutex
	tableName  string
}

// New returns a new Darwinx struct
func New(pool *pgxpool.Pool, options ...Option) (*Darwinx, error) {
	d := &Darwinx{
		pool:      pool,
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
	return d.run(ctx, records)
}

func (d *Darwinx) applyMigration(ctx context.Context, migration Migration) error {
	start := time.Now()

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	committed := false
	defer func(tx pgx.Tx) {
		if !committed {
			_ = tx.Rollback(context.Background())
		}
	}(tx)

	if _, err := tx.Exec(ctx, migration.Script); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("migration %s failed during Exec", formatVersion(migration.Version)))
	}

	rec := MigrationRecord{
		Version:       migration.Version,
		Description:   migration.Description,
		Checksum:      migration.Checksum(),
		AppliedAt:     start,
		ExecutionTime: time.Since(start),
	}
	if err := d.insertRecord(ctx, tx, rec); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("migration %s failed during record insert", formatVersion(migration.Version)))
	}

	if err := tx.Commit(ctx); err != nil {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil {
			return errors.Wrapf(err, "commit failed for migration %s; rollback also failed: %v",
				formatVersion(migration.Version), rbErr)
		}
		return errors.WithMessage(err, fmt.Sprintf("commit failed for migration %s; rolled back", formatVersion(migration.Version)))
	}
	return nil
}

func (d *Darwinx) run(ctx context.Context, records []MigrationRecord) error {
	migrations := planMigration(records, d.migrations)
	for _, migration := range migrations {
		if err := d.applyMigration(ctx, migration); err != nil {
			return err
		}
	}
	return nil
}

func formatVersion(v float64) string {
	return fmt.Sprintf("%.3f", v)
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

// Records returns all migration records from the database in descending order
func (d *Darwinx) Records(ctx context.Context) ([]MigrationRecord, error) {
	records, err := d.allFromDB(ctx)
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(byMigrationRecordVersion(records)))

	return records, nil
}

func (d *Darwinx) allFromDB(ctx context.Context) ([]MigrationRecord, error) {
	query := fmt.Sprintf(`SELECT
	version,
	description,
	checksum,
	applied_at,
	execution_time
FROM %s
	ORDER BY version ASC;`, d.tableName)
	rows, err := d.pool.Query(ctx, query)
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
	_, err := d.pool.Exec(ctx, fmt.Sprintf(schemaQuery, d.tableName))
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
	if err != nil {
		return errors.WithMessage(err, "insert record failed")
	}
	return nil
}

func planMigration(records []MigrationRecord, migrations []Migration) []Migration {
	if len(records) == 0 {
		sort.Sort(byMigrationVersion(migrations))
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
