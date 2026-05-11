package darwinx

import (
	"context"
	"sort"
)

// Validate if the database migrations are applied and consistent
func (d *Darwinx) Validate(ctx context.Context) error {
	applied, err := d.allFromDB(ctx)
	if err != nil {
		return err
	}
	return d.validateWithRecords(applied)
}

// validateWithRecords runs static checks and compares the given DB snapshot to d.migrations.
func (d *Darwinx) validateWithRecords(applied []MigrationRecord) error {
	for _, migration := range d.migrations {
		if err := validateVersion(migration); err != nil {
			return err
		}
	}

	if err := validateDuplicates(d.migrations); err != nil {
		return err
	}

	if removed := getRemovedMigration(applied, d.migrations); len(removed) != 0 {
		return RemovedMigrationsError{
			Versions: removed,
		}
	}

	if changed := compareMigrationChecksums(applied, d.migrations); len(changed) != 0 {
		return InvalidChecksumError{
			Versions: changed,
		}
	}
	return nil
}

func getRemovedMigration(applied []MigrationRecord, migrations []Migration) []float64 {
	versionMap := map[float64]struct{}{}
	var removed []float64

	for _, migration := range migrations {
		versionMap[migration.Version] = struct{}{}
	}

	for _, migration := range applied {
		if _, ok := versionMap[migration.Version]; !ok {
			removed = append(removed, migration.Version)
		}
	}

	return removed
}

func compareMigrationChecksums(applied []MigrationRecord, migrations []Migration) []float64 {
	versionMap := map[float64]MigrationRecord{}

	for _, migration := range applied {
		versionMap[migration.Version] = migration
	}

	var changed []float64
	for _, migration := range migrations {
		if m, ok := versionMap[migration.Version]; ok {
			if m.Checksum != migration.Checksum() {
				changed = append(changed, migration.Version)
			}
		}
	}
	sort.Float64s(changed)
	return changed
}

func validateVersion(m Migration) error {
	if m.Version < 0 {
		return IllegalMigrationVersionError{Version: m.Version}
	}
	return nil
}

func validateDuplicates(migrations []Migration) error {
	unique := map[float64]struct{}{}
	for _, migration := range migrations {
		_, exists := unique[migration.Version]
		if exists {
			return DuplicateMigrationVersionError{Version: migration.Version}
		}
		unique[migration.Version] = struct{}{}
	}
	return nil
}
