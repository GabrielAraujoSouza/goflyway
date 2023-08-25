package goflyway

import "errors"

var (
	ErrDatabaseConnectionNull    = errors.New("database connection is null")
	ErrUnsupportedDatabaseDriver = errors.New("unsupported database driver")
	ErrRunnerNotInitialized      = errors.New("runner not initialized")
	ErrLocationCannotBeEmpty     = errors.New("migration location cannot be empty")
)

var (
	WarnNoMigrationFound = "no migrations found, are your location set up correctly?"
)
