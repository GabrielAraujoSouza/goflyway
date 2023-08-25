package goflyway

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

type historyModelDb struct {
	InstalledRank *int
	Version       *string
	Description   *string
	Type          *string
	Script        *string
	Checksum      *string
	InstalledBy   *string
	InstalledOn   *string
	ExecutionTime *int
	Success       *bool
}

var fail = func(err error) (int64, error) {
	return 0, fmt.Errorf("error inserting migration history: %v", err)
}

func getCreateTableCommand(driver driver, tableName string) string {
	var createCommand string

	switch driver {
	case POSTGRES:
		createCommand = createTablePostgres
	case MYSQL:
		createCommand = createTableMysql
	case MSSQLSERVER:
		createCommand = createTableMsSqlServer
	}

	return regexTableName.ReplaceAllString(createCommand, tableName)
}

func getSelectTableCommand(driver driver, tableName string) string {
	var selectCommand string

	switch driver {
	case POSTGRES:
		selectCommand = selectTablePostgres
	case MYSQL:
		selectCommand = selectTableMysql
	case MSSQLSERVER:
		selectCommand = selectTableMsSqlServer
	}
	return regexTableName.ReplaceAllString(selectCommand, tableName)
}

func parseInsertMigration(driver driver, tableName string) string {
	var insertCommand string

	switch driver {
	case POSTGRES:
		insertCommand = insertPostgres
	case MYSQL:
		insertCommand = insertMysql
	case MSSQLSERVER:
		insertCommand = insertMsSqlServer
	}
	return regexTableName.ReplaceAllString(insertCommand, tableName)
}

func validateDriver(s string) error {
	switch s {
	case string(POSTGRES), string(MYSQL), string(MSSQLSERVER):
		return nil
	default:
		return ErrUnsupportedDatabaseDriver
	}
}

// selectMigrationHistory Query migration table
func selectMigrationHistory(db *sql.DB, query string) ([]historyModel, error) {
	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var migrations []historyModelDb

	for rows.Next() {
		var m historyModelDb
		err = rows.Scan(
			&m.InstalledRank,
			&m.Version,
			&m.Description,
			&m.Type,
			&m.Script,
			&m.Checksum,
			&m.InstalledBy,
			&m.InstalledOn,
			&m.ExecutionTime,
			&m.Success,
		)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	var result []historyModel

	for _, v := range migrations {
		var m historyModel

		if v.InstalledRank != nil {
			m.InstalledRank = *v.InstalledRank
		}
		if v.Version != nil {
			m.Version = *v.Version
		}
		if v.Description != nil {
			m.Description = *v.Description
		}
		if v.Type != nil {
			m.Type = *v.Type
		}
		if v.Script != nil {
			m.Script = *v.Script
		}
		if v.Checksum != nil {
			m.Checksum = *v.Checksum
		}
		if v.InstalledBy != nil {
			m.InstalledBy = *v.InstalledBy
		}

		if v.InstalledOn != nil {
			t, err := time.Parse("2006-01-02T15:04:05Z", *v.InstalledOn)
			if err != nil {
				printWarningLog(fmt.Sprintf("error parse installed_on: %v", err))
			} else {
				m.InstalledOn = &t
			}
		}

		if v.ExecutionTime != nil {
			m.ExecutionTime = *v.ExecutionTime
		}
		if v.Success != nil {
			m.Success = *v.Success
		}

		result = append(result, m)
	}

	return result, nil
}

func executeMigration(db *sql.DB, insertQuery string, history historyModel, g *goFlywayRunner) (*historyModel, error) {

	startExec := time.Now().UnixMilli()

	// execute
	_, err := executeScript(db, history, g)
	if err != nil {
		return nil, err
	}
	endExec := time.Now().UnixMilli()

	total := int(endExec - startExec)
	history.ExecutionTime = total

	_, err = insertMigration(db, insertQuery, history, g)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func insertMigration(db *sql.DB, insertQuery string, history historyModel, g *goFlywayRunner) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	r1, err := insertExecutor(tx, insertQuery, history, g)
	if err != nil {
		return fail(err)
	}

	rw, err := r1.RowsAffected()
	if err != nil {
		return fail(err)
	}
	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fail(err)
	}

	return rw, nil
}

func executeScript(db *sql.DB, history historyModel, g *goFlywayRunner) (int64, error) {

	scriptFile := fmt.Sprintf("%s/%s", g.config.Location, history.Script)

	b, err := os.ReadFile(scriptFile)
	if err != nil {
		return 0, err
	}

	query := string(b)

	tx, err := db.Begin()
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	// r1, err := tx.Exec(query)
	r1, err := queryExecutor(tx, query, g)
	if err != nil {
		return fail(err)
	}

	rw, err := r1.RowsAffected()
	if err != nil {
		printWarningLog(fmt.Sprintf("warning: %v", err))
	}
	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fail(err)
	}

	return rw, nil
}

func queryExecutor(tx *sql.Tx, query string, g *goFlywayRunner) (sql.Result, error) {

	if g.config.Driver == POSTGRES {
		return tx.Exec(query)
	}

	if g.config.Driver == MYSQL {
		// at moment it's not possible to create migration with multiple statements
		return tx.Exec(query)
	}

	if g.config.Driver == MSSQLSERVER {
		return tx.Exec(query)
	}

	return nil, fmt.Errorf("no driver found on execute")
}

func insertExecutor(tx *sql.Tx, insertQuery string, history historyModel, g *goFlywayRunner) (sql.Result, error) {

	if g.config.Driver == MSSQLSERVER {

		return tx.Exec(insertQuery,
			sql.Named("installed_rank", history.InstalledRank),
			sql.Named("version", history.Version),
			sql.Named("description", history.Description),
			sql.Named("type", history.Type),
			sql.Named("script", history.Script),
			sql.Named("checksum", history.Checksum),
			sql.Named("execution_time", history.ExecutionTime))
	}

	return tx.Exec(insertQuery,
		history.InstalledRank,
		history.Version,
		history.Description,
		history.Type,
		history.Script,
		history.Checksum,
		history.ExecutionTime)
}
