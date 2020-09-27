package darwinx

// Status is a migration status value
type Status uint8

const (
	// Ignored means that the migrations was not applied to the database
	Ignored Status = iota
	// Applied means that the migrations was successfully applied to the database
	Applied
	// Pending means that the migrations is a new migration and it is waiting to be applied to the database
	Pending
	// Error means that the migration could not be applied to the database
	Error
)

func (s Status) String() string {
	switch s {
	case Ignored:
		return "IGNORED"
	case Applied:
		return "APPLIED"
	case Pending:
		return "PENDING"
	case Error:
		return "ERROR"
	default:
		return "INVALID"
	}
}

func getStatus(inDatabase []MigrationRecord, migration Migration) Status {
	if len(inDatabase) == 0 {
		return Pending
	}
	last := inDatabase[0]

	if migration.Version > last.Version {
		return Pending
	}

	found := false
	for _, record := range inDatabase {
		if record.Version == migration.Version {
			found = true
			break
		}
	}

	if !found {
		return Ignored
	}

	return Applied
}
