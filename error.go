package darwinx

import (
	"fmt"
	"strings"
)

// DuplicateMigrationVersionError is used to report when the migration list has duplicated entries
type DuplicateMigrationVersionError struct {
	Version float64
}

func (d DuplicateMigrationVersionError) Error() string {
	return fmt.Sprintf("Multiple migrations have the version number %f.", d.Version)
}

// IllegalMigrationVersionError is used to report when the migration has an illegal Version number
type IllegalMigrationVersionError struct {
	Version float64
}

func (i IllegalMigrationVersionError) Error() string {
	return fmt.Sprintf("Illegal migration version number %f.", i.Version)
}

// RemovedMigrationError is used to report when a migration is removed from the list
type RemovedMigrationsError struct {
	Versions []float64
}

func (r RemovedMigrationsError) Error() string {
	var parts []string
	for _, version := range r.Versions {
		parts = append(parts, fmt.Sprintf("%f", version))
	}
	return fmt.Sprintf("Migration %s were removed", strings.Join(parts, ", "))
}

// InvalidChecksumError is used to report when a migration was modified
type InvalidChecksumError struct {
	Version float64
}

func (i InvalidChecksumError) Error() string {
	return fmt.Sprintf("Invalid cheksum for migration %f", i.Version)
}
