# Go Flyway

<img align="right" width="159px" src="https://raw.githubusercontent.com/gabrielaraujosouza/logo/master/goflyway-logo.png">

Database migration library written in go inspired by <a href="https://flywaydb.org/" target="_blank">flyway</a>

## Download

```shell
go get github.com/gabrielaraujosouza/goflyway
```


## Key Validations
### Duplicated version
- It checks if there are scritps with same version number
### Checksum mismatch
- This guarantees the integrity of the script, that is, it is not possible to edit a script after it has been executed

## Supported Databases
- PostgreSQL
- MySQL (connection instance must have `multiStatements` set to `true`)
- Microsoft SQL Server
- Sqlite3

## Usage

```go
import (
	"database/sql"
	"fmt"

	"github.com/gabrielaraujosouza/goflyway"
	_ "github.com/lib/pq"
)

func main() {

	dbUrl := "postgres://root:root@localhost:5432/goflyway?sslmode=disable"

	// you need to open a db connection
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// create a config
	conf := goflyway.GoFlywayConfig{
		Db:       db,
		Driver:   goflyway.POSTGRES,
		Location: "[SCRIPT_FOLDER_PATH_HERE]", // Example: /home/user/my-prokect/db/migration
	}

	// Call the method Migrate
	totalScriptsExecuted, err := goflyway.Migrate(conf)

	if err != nil {
		panic(err)
	}

	fmt.Println("total migrations applied:", totalScriptsExecuted)
}

```
Output:
```
successfully validated 3 migrations (execution time 0ms)
current version of schema: << Empty Schema >>
migrating schema to version 1 - test create table product
migrating schema to version 2 - test alter table product
migrating schema to version 3 - test remove column from product
successfully applied 3 migrations to schema, now at version v3 (execution time 146ms)

```

## Config Properties

Property | Default | Description |
--------|------------|--------
**Table** | `goflyway_schema_history` | `Name of the schema history table that will be used by GoFlyway` 
**SqlMigrationPrefix** | `V` | `File name prefix for SQL migrations` | Used for stable releases |
**SqlMigrationSeparator** | `__` | `File name separator for SQL migrations.`
**Location** | - | `Location of migrations scripts`
**OutOfOrder** | `false` |`Whether to allow migrations to be run out of order`
**IgnoreMissingMigrations** | `false` | `Ignore missing migrations`
**Db** | -| `Database connection`
**Driver** | - | `Database drive`
**ShowWarningLog** | `false`| `Shows warning logs`

