package darwinx

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// Migration represents a database migrations.
type Migration struct {
	Version     float64
	Description string
	Script      string
}

// MigrationRecord is the entry in schema table
type MigrationRecord struct {
	Version       float64
	Description   string
	Checksum      string
	AppliedAt     time.Time
	ExecutionTime time.Duration
}


// Checksum calculate the Script sha256
func (m Migration) Checksum() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(m.Script)))
}

// MigrationInfo is a struct used in the infoChan to inform clients about
// the migration being applied.
type MigrationInfo struct {
	Status    Status
	Error     error
	Migration Migration
}
