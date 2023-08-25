package goflyway

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type historyModel struct {
	InstalledRank int
	Version       string
	Description   string
	Type          string
	Script        string
	Checksum      string
	InstalledBy   string
	InstalledOn   *time.Time
	ExecutionTime int
	Success       bool
}

type localScript struct {
	Version     string
	Description string
	Script      string
	Checksum    string
}

type GoFlywayConfig struct {
	// Name of the schema history table that will be used by GoFlyway. Defaul is "goflyway_schema_history"
	Table string

	// File name prefix for SQL migrations. Default is "V"
	SqlMigrationPrefix string

	// File name separator for SQL migrations. Default is "__"
	SqlMigrationSeparator string

	// Location of migrations scripts. Examle: "/home/user/my-project/migrations"
	Location string

	// Whether to allow migrations to be run out of order. Default is "false"
	OutOfOrder bool

	// Ignore missing migrations. Default is "false"
	IgnoreMissingMigrations bool

	// Database connection
	Db *sql.DB

	// Database drive
	Driver driver

	// Shows warning logs. Default is "false"
	ShowWarningLog bool

	// File name sufix for SQL migrations. Default is ".sql"
	sqlMigrationSuffix string
}

type goFlywayRunner struct {
	config      GoFlywayConfig
	initialized bool
}

// Migrate apply migrations to database and returns the total of executed migrations
func Migrate(c GoFlywayConfig) (int, error) {

	g, err := newGoFlywayRunner(c)
	if err != nil {
		return 0, err
	}

	if !g.initialized {
		return 0, ErrRunnerNotInitialized
	}

	mFiles, err := g.readLocalMigrations()
	if err != nil {
		return 0, err
	}

	mTable, err := g.readMigrationTable()
	if err != nil {
		return 0, err
	}

	err = g.validateMigrations(mFiles, mTable)
	if err != nil {
		return 0, err
	}

	total, err := g.applyMigrations(mFiles, mTable)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// CalculateChecksum generate file checksum, it is used to check script integrity
func CalculateChecksum(filename string) (string, error) {

	f, err := os.Open(filename)
	if err != nil {
		return "", throwErrMigration(err)
	}
	defer f.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, f)
	if err != nil {
		return "", throwErrMigration(err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func newGoFlywayRunner(config GoFlywayConfig) (*goFlywayRunner, error) {

	c := &goFlywayRunner{
		config: config,
	}

	err := c.applyDefaultSettings()
	if err != nil {
		return nil, err
	}

	c.initialized = true

	return c, nil
}

// applyDefaultSettings Fill config values if it is empty
func (g *goFlywayRunner) applyDefaultSettings() error {

	g.config.sqlMigrationSuffix = ".sql"

	if len(g.config.Table) <= 0 {
		g.config.Table = tableName
	}

	if len(g.config.SqlMigrationPrefix) <= 0 {
		g.config.SqlMigrationPrefix = sqlMigrationPrefix
	}

	if len(g.config.SqlMigrationSeparator) <= 0 {
		g.config.SqlMigrationSeparator = sqlMigrationSeparator
	}

	if len(g.config.Location) <= 0 {
		return ErrLocationCannotBeEmpty
	}

	err := validateDriver(string(g.config.Driver))
	if err != nil {
		return err
	}

	showWarningLog = g.config.ShowWarningLog

	return nil
}

// ReadLocalMigrations Load migration files
func (g *goFlywayRunner) readLocalMigrations() ([]localScript, error) {

	fail := func(err error) ([]localScript, error) {
		return nil, throwErrMigration(fmt.Errorf("error reading local migrations: %v", err))
	}

	c := g.config
	sqlFiles := []localScript{}

	migrationDir, err := os.ReadDir(c.Location)
	if err != nil {
		return fail(err)
	}

	if len(migrationDir) <= 0 {
		printWarningLog(warnNoMigrationFound)
	}

	for _, f := range migrationDir {

		if strings.HasPrefix(f.Name(), c.SqlMigrationPrefix) && strings.HasSuffix(f.Name(), c.sqlMigrationSuffix) {
			version, description, err := extractValuesFromScriptName(
				f.Name(), g.config.SqlMigrationPrefix, g.config.SqlMigrationSeparator, g.config.sqlMigrationSuffix)

			if err != nil {
				printWarningLog(fmt.Sprintf("warning: %v", err))
			} else {
				sf := localScript{
					Version:     version,
					Description: description,
					Script:      f.Name(),
				}

				cSfileCheckSumum, err := CalculateChecksum(fmt.Sprintf("%s/%s", c.Location, sf.Script))
				if err != nil {
					printWarningLog(fmt.Sprintf("warning: checksum calculation error for script %s: %v ", sf.Script, err))
				} else {
					sf.Checksum = cSfileCheckSumum
					sqlFiles = append(sqlFiles, sf)
				}
			}
		}
	}

	if len(sqlFiles) <= 0 {
		printWarningLog(warnNoMigrationFound)
	}

	sort.SliceStable(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Version < sqlFiles[j].Version
	})

	return sqlFiles, nil
}

// ReadMigrationTable Load database migrations
func (g *goFlywayRunner) readMigrationTable() ([]historyModel, error) {

	fail := func(err error) ([]historyModel, error) {
		return nil, throwErrMigration(fmt.Errorf("error reading migration table: %v", err))
	}

	db := g.config.Db
	if db == nil {
		return fail(ErrDatabaseConnectionNull)
	}

	tableValue := g.config.Table

	// always try to create history table to evict errors
	queryCreateTable := getCreateTableCommand(g.config.Driver, g.config.Table)

	_, err := db.Exec(queryCreateTable)
	if err != nil {
		return fail(err)
	}

	// list  migrations
	queryTable := getSelectTableCommand(g.config.Driver, tableValue)
	migrations, err := selectMigrationHistory(g.config.Db, queryTable)
	if err != nil {
		return fail(err)
	}

	return migrations, nil
}

func (g *goFlywayRunner) validateMigrations(localMigrations []localScript, databaseMigrations []historyModel) error {

	startExec := time.Now().UnixMilli()

	// validate local migrations
	for i, lm := range localMigrations {

		// check if local migrations has duplicated version
		dupLocalMg := findLocalMigrationsByVersion(localMigrations, lm.Version)

		if len(dupLocalMg) > 1 {
			return throwErrMigration(fmt.Errorf("found more than one migration with version %s: %s",
				lm.Version, getScriptNames(dupLocalMg)))
		}

		dm := findMigrationByVersion(databaseMigrations, lm.Version)

		if dm != nil {

			if dm.Checksum != lm.Checksum {
				return throwErrMigration(fmt.Errorf("migration checksum mismatch for migration version %s: applied to database = %s, resolved locally = %s",
					lm.Version, dm.Checksum, lm.Checksum))
			}

			if dm.Description != lm.Description {
				return throwErrMigration(fmt.Errorf("migration description mismatch for migration version %s: applied to database = %s, resolved locally = %s",
					lm.Version, dm.Description, lm.Description))
			}
		}

		// check if is out of order
		if !g.config.OutOfOrder {

			migrationIndex := findMigrationIndexByVersion(databaseMigrations, lm.Version)
			if migrationIndex == -1 && i < len(databaseMigrations) {
				return throwErrMigration(fmt.Errorf("detected resolved migration not applied to database: %s, to allow executing this migration, set OutOfOrder=true",
					lm.Version))
			}
		}
	}

	// validate database migrations
	for _, dm := range databaseMigrations {

		// check if any applied migration is missing
		if !g.config.IgnoreMissingMigrations {

			lm := findLocalMigrationByVersion(localMigrations, dm.Version)

			if lm == nil {
				return throwErrMigration(fmt.Errorf("detected applied migration not resolved locally: %s", dm.Version))
			}
		}
	}

	endExec := time.Now().UnixMilli()

	executionTime := int(endExec - startExec)

	logg.Printf("successfully validated %d migrations (execution time %dms)",
		len(localMigrations), executionTime) // TODO format to time

	return nil
}

func (gr *goFlywayRunner) applyMigrations(localMigrations []localScript, databaseMigrations []historyModel) (int, error) {

	startExec := time.Now().UnixMilli()

	countMigrations := 0
	executedMigrations := databaseMigrations

	installedRank := findLargestInstalledRank(executedMigrations)

	var latestVersion string
	if len(databaseMigrations) > 0 {
		latestVersion = databaseMigrations[len(databaseMigrations)-1].Version
	}

	if len(latestVersion) == 0 {
		logg.Printf("current version of schema: << Empty Schema >>")
	} else {
		logg.Printf("current version of schema: %s", latestVersion)
	}

	for _, lm := range localMigrations {

		migrationExecuted := findMigrationByVersion(executedMigrations, lm.Version)

		if migrationExecuted == nil {

			installedRank++

			newMigration := historyModel{
				Version:       lm.Version,
				Description:   lm.Description,
				Script:        lm.Script,
				Type:          "sql",
				Checksum:      lm.Checksum,
				InstalledRank: installedRank,
			}

			_, err := executeMigration(gr.config.Db, parseInsertMigration(gr.config.Driver, gr.config.Table), newMigration, gr)
			if err != nil {
				return countMigrations, throwErrMigration(fmt.Errorf("migration %s failed: %v", newMigration.Script, err))
			}

			logg.Printf("migrating schema to version %s - %s", newMigration.Version, newMigration.Description)

			executedMigrations = append(executedMigrations, newMigration)
			countMigrations++

			latestVersion = newMigration.Version
		}
	}

	endExec := time.Now().UnixMilli()
	executionTime := int(endExec - startExec)

	if countMigrations == 0 {
		logg.Printf("schema is up to date, no migration necessary")
	} else {
		logg.Printf("successfully applied %d migrations to schema, now at version v%s (execution time %dms)",
			countMigrations, latestVersion, executionTime) // TODO format to time
	}

	return countMigrations, nil
}
