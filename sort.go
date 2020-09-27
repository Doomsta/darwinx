package darwinx

type byMigrationVersion []Migration
func (b byMigrationVersion) Len() int           { return len(b) }
func (b byMigrationVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byMigrationVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }

type byMigrationRecordVersion []MigrationRecord
func (b byMigrationRecordVersion) Len() int           { return len(b) }
func (b byMigrationRecordVersion) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byMigrationRecordVersion) Less(i, j int) bool { return b[i].Version < b[j].Version }
