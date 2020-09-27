package darwinx

import "context"

// Validate if the database migrations are applied and consistent
func (d *Darwinx) Validate(ctx context.Context) error {
	for _, migration := range d.migrations {
		if err := validateVersion(migration); err != nil {
			return err
		}
	}

	if err := validateDuplicates(d.migrations); err != nil {
		return err
	}

	applied, err := d.allFromDB(ctx)
	if err != nil {
		return err
	}

	if removed := getRemovedMigration(applied, d.migrations); len(removed) != 0 {
		return RemovedMigrationsError{
			Versions: removed,
		}
	}

	if err := compareMigrationChecksums(applied, d.migrations); err != nil {
		return err
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

func compareMigrationChecksums(applied []MigrationRecord, migrations []Migration) error {
	versionMap := map[float64]MigrationRecord{}

	for _, migration := range applied {
		versionMap[migration.Version] = migration
	}

	for _, migration := range migrations {
		if m, ok := versionMap[migration.Version]; ok {
			if m.Checksum != migration.Checksum() {
				return InvalidChecksumError{
					Version: migration.Version,
				}
			}
		}
	}

	return nil
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
