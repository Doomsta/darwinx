package darwinx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareMigrationChecksums_collectsAllMismatches(t *testing.T) {
	m1 := Migration{Version: 1, Description: "a", Script: "SELECT 1"}
	m2 := Migration{Version: 2, Description: "b", Script: "SELECT 2"}
	applied := []MigrationRecord{
		{Version: 1, Checksum: m1.Checksum()},
		{Version: 2, Checksum: m2.Checksum()},
	}
	altered := []Migration{
		{Version: 1, Description: "a", Script: "SELECT 1 changed"},
		{Version: 2, Description: "b", Script: "SELECT 2 changed"},
	}
	changed := compareMigrationChecksums(applied, altered)
	assert.Equal(t, []float64{1, 2}, changed)
}
