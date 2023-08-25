package goflyway

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const tableName = "goflyway_schema_history"
const sqlMigrationPrefix = "V"
const sqlMigrationSeparator = "__"

var logg = log.Default()

var regexTableName = regexp.MustCompile(`\[tableName\]`)
var regexVersion = regexp.MustCompile(`^\d((_\d)|(\d))*$`)
var showWarningLog bool

type driver string

const (
	POSTGRES driver = "postgres"
	MYSQL    driver = "mysql"
)

func extractValuesFromScriptName(name string, prefix string, separator string, sufix string) (string, string, error) {

	if !strings.Contains(name, separator) {
		return "", "", fmt.Errorf("migration '%s' does not contains separator '%s'", name, separator)
	}

	version := name[len(prefix):strings.Index(name, separator)]
	if len(version) <= 0 || !regexVersion.MatchString(version) {
		return "", "", fmt.Errorf("invalid version '%s' for migration '%s'", version, name)
	}
	version = strings.ReplaceAll(version, "_", ".")

	description := name[strings.Index(name, separator)+len(separator) : strings.LastIndex(name, sufix)]

	if len(description) <= 0 {
		return "", "", fmt.Errorf("migration description cannot be empty: '%s'", name)
	}

	return version, strings.TrimSpace(strings.ReplaceAll(description, "_", " ")), nil
}

func findMigrationByVersion(migrations []historyModel, version string) *historyModel {
	for _, m := range migrations {
		if m.Version == version {
			return &m
		}
	}
	return nil
}

func findMigrationIndexByVersion(migrations []historyModel, version string) int {
	for i, m := range migrations {
		if m.Version == version {
			return i
		}
	}
	return -1
}

func findLocalMigrationsByVersion(migrations []localScript, version string) []localScript {
	mg := []localScript{}
	for _, m := range migrations {
		if m.Version == version {
			mg = append(mg, m)
		}
	}
	return mg
}

func findLocalMigrationByVersion(migrations []localScript, version string) *localScript {
	for _, m := range migrations {
		if m.Version == version {
			return &m
		}
	}
	return nil
}

func findLargestInstalledRank(migrations []historyModel) int {

	largest := 0
	for _, m := range migrations {
		if m.InstalledRank > largest {
			largest = m.InstalledRank
		}
	}

	return largest
}

func getScriptNames(localMigrations []localScript) string {
	s := []string{}
	for _, m := range localMigrations {
		s = append(s, m.Script)
	}

	return strings.Join(s, " ")
}

func printWarningLog(message string) {

	if showWarningLog {
		logg.Println(message)
	}
}

func getWorkPath() string {
	_, b, _, _ := runtime.Caller(0)

	return filepath.Dir(b)
}
