package goflyway

import (
	"fmt"
	"testing"
	"time"
)

func TestValidateMigrations(t *testing.T) {

	g, err := newGoFlywayRunner(GoFlywayConfig{
		Driver:   POSTGRES,
		Location: getWorkPath() + "/utils/test/db/custom-migration",
	})

	if err != nil {
		t.Fatalf("errors happened when initialize goflywayrunner: %v", err)
	}

	dbMigrations := getDatabaseMigrations()

	localMigrations := getLocalMigrations()

	err = g.validateMigrations(localMigrations, dbMigrations)

	if err != nil {
		t.Errorf("expected nil but got %v", err)
	}
}

func TestValidateMigrations_DuplicatedVersionError(t *testing.T) {

	g, err := newGoFlywayRunner(GoFlywayConfig{
		Driver:   POSTGRES,
		Location: getWorkPath() + "/utils/test/db/custom-migration",
	})

	if err != nil {
		t.Fatalf("errors happened when initialize goflywayrunner: %v", err)
	}

	expectedMessage := fmt.Sprintf("found more than one migration with version %s: %s",
		"2", "V2__test_alter_table_product.sql V2__test_duplicated_alter_table_product.sql")

	dbMigrations := getDatabaseMigrations()

	localMigrations := getLocalMigrations()
	localMigrations = append(localMigrations, localScript{
		Version:     "2",
		Description: "test duplicated alter table product",
		Script:      "V2__test_duplicated_alter_table_product.sql",
		Checksum:    "2eaef764b14c8a99535a61a9f4fd4af9428e1905bfa68659e6af2ed2c45d3a02",
	})

	err = g.validateMigrations(localMigrations, dbMigrations)

	if err == nil {
		t.Errorf("expected error %v but got nil", expectedMessage)
	}

	if err.Error() != expectedMessage {
		t.Errorf("expected error %v but got %v", expectedMessage, err.Error())
	}
}

func TestValidateMigrations_ChecksumMismatchError(t *testing.T) {

	g, err := newGoFlywayRunner(GoFlywayConfig{
		Driver:   POSTGRES,
		Location: getWorkPath() + "/utils/test/db/custom-migration",
	})

	if err != nil {
		t.Fatalf("errors happened when initialize goflywayrunner: %v", err)
	}

	expectedMessage := fmt.Sprintf("migration checksum mismatch for migration version %s: applied to database = %s, resolved locally = %s",
		"2", "2eaef764b14c8a99535a61a9f4fd4af9428e1905bfa68659e6af2ed2c45d3a02", "ad237d5f6002d5dbad359f98e2ef38dc74bb0e4eb838dafe923cfe7229b5024c")

	dbMigrations := getDatabaseMigrations()

	localMigrations := getLocalMigrations()
	localMigrations[1].Checksum = "ad237d5f6002d5dbad359f98e2ef38dc74bb0e4eb838dafe923cfe7229b5024c"

	err = g.validateMigrations(localMigrations, dbMigrations)

	if err == nil {
		t.Errorf("expected error %v but got nil", expectedMessage)
	}

	if err.Error() != expectedMessage {
		t.Errorf("expected error %v but got %v", expectedMessage, err.Error())
	}
}

func TestValidateMigrations_DescriptionMismatchError(t *testing.T) {

	g, err := newGoFlywayRunner(GoFlywayConfig{
		Driver:   POSTGRES,
		Location: getWorkPath() + "/utils/test/db/custom-migration",
	})

	if err != nil {
		t.Fatalf("errors happened when initialize goflywayrunner: %v", err)
	}

	expectedMessage := fmt.Sprintf("migration description mismatch for migration version %s: applied to database = %s, resolved locally = %s",
		"2", "test alter table product", "test update table product")

	dbMigrations := getDatabaseMigrations()

	localMigrations := getLocalMigrations()
	localMigrations[1].Description = "test update table product"

	err = g.validateMigrations(localMigrations, dbMigrations)

	if err == nil {
		t.Errorf("expected error %v but got nil", expectedMessage)
	}

	if err.Error() != expectedMessage {
		t.Errorf("expected error %v but got %v", expectedMessage, err.Error())
	}
}

func TestValidateMigrations_OutOfOrderError(t *testing.T) {

	g, err := newGoFlywayRunner(GoFlywayConfig{
		Driver:   POSTGRES,
		Location: getWorkPath() + "/utils/test/db/custom-migration",
	})

	if err != nil {
		t.Fatalf("errors happened when initialize goflywayrunner: %v", err)
	}

	expectedMessage := fmt.Sprintf("detected resolved migration not applied to database: %s, to allow executing this migration, set OutOfOrder=true", "2")

	dbMigrations := getDatabaseMigrations()

	// removing db migration version 2
	elemntIndex := 1
	copy(dbMigrations[elemntIndex:], dbMigrations[elemntIndex+1:])
	dbMigrations = dbMigrations[:len(dbMigrations)-1]

	localMigrations := getLocalMigrations()

	err = g.validateMigrations(localMigrations, dbMigrations)

	if err == nil {
		t.Errorf("expected error %v but got nil", expectedMessage)
	}

	if err.Error() != expectedMessage {
		t.Errorf("expected error %v but got %v", expectedMessage, err.Error())
	}
}

func TestValidateMigrations_ValidateMissionMigrationError(t *testing.T) {

	g, err := newGoFlywayRunner(GoFlywayConfig{
		Driver:   POSTGRES,
		Location: getWorkPath() + "/utils/test/db/custom-migration",
	})

	if err != nil {
		t.Fatalf("errors happened when initialize goflywayrunner: %v", err)
	}

	expectedMessage := fmt.Sprintf("detected applied migration not resolved locally: %s", "2")

	dbMigrations := getDatabaseMigrations()

	localMigrations := getLocalMigrations()

	// removing local migration version 2
	elemntIndex := 1
	copy(localMigrations[elemntIndex:], localMigrations[elemntIndex+1:])
	localMigrations = localMigrations[:len(localMigrations)-1]

	err = g.validateMigrations(localMigrations, dbMigrations)

	if err == nil {
		t.Errorf("expected error %v but got nil", expectedMessage)
	}

	if err.Error() != expectedMessage {
		t.Errorf("expected error %v but got %v", expectedMessage, err.Error())
	}
}

func TestReadLocalMigrations(t *testing.T) {

	location := getWorkPath() + "/utils/test/db/migration/postgres"
	// using default config values
	g, err := newGoFlywayRunner(GoFlywayConfig{
		Driver:   POSTGRES,
		Location: location,
	})

	if err != nil {
		t.Fatalf("errors happened when initialize goflywayrunner: %v", err)
	}

	ignoredMigrations := []string{
		"V4_invalid_separtor.sql",
		"V5a__invalid_version.sql",
		"V7__.sql",
		"V8__invalid_suffix.txt",
		"W6__invalid_prefix.sql",
	}

	localMigrations, err := g.readLocalMigrations()

	if err != nil {
		t.Fatalf("expected nil but got error %v", err)
	}

	if len(localMigrations) <= 0 {
		t.Errorf("expected migrations length greater than 0 but got empty migrations")
	}

	for _, l := range localMigrations {

		for _, im := range ignoredMigrations {
			if l.Script == im {
				t.Errorf("expected migration should not exists, but got %s", im)
			}
		}
	}
}

func TestReadLocalMigrations_UsingCustomMigrationPattern(t *testing.T) {

	location := getWorkPath() + "/utils/test/db/custom-migration"
	// using default config values
	g, err := newGoFlywayRunner(GoFlywayConfig{
		Driver:                POSTGRES,
		Location:              location,
		SqlMigrationPrefix:    "GV",
		SqlMigrationSeparator: "--",
	})

	if err != nil {
		t.Fatalf("errors happened when initialize goflywayrunner: %v", err)
	}

	ignoredMigrations := []string{
		"GV4_invalid_separtor.sql",
		"GV5a--invalid_version.sql",
		"GV7--.sql",
		"GV8--invalid_suffix.txt",
		"GW6--invalid_prefix.sql",
	}

	localMigrations, err := g.readLocalMigrations()

	if err != nil {
		t.Fatalf("expected nil but got error %v", err)
	}

	if len(localMigrations) <= 0 {
		t.Errorf("expected migrations length greater than 0 but got empty migrations")
	}

	for _, l := range localMigrations {

		for _, im := range ignoredMigrations {
			if l.Script == im {
				t.Errorf("expected migration should not exists, but got %s", im)
			}
		}
	}
}

func TestCalculateChecksum(t *testing.T) {

	type ChecksumExpected struct {
		Filename         string
		ExpectedChecksum string
	}

	fileList := []ChecksumExpected{
		{
			Filename:         "V1__test_create_table_product.sql",
			ExpectedChecksum: "423156a418cd885a0304a6f0cc2ad8e7059e8421ae181800257fa31815e9d197",
		},
		{
			Filename:         "V2__test_alter_table_product.sql",
			ExpectedChecksum: "2eaef764b14c8a99535a61a9f4fd4af9428e1905bfa68659e6af2ed2c45d3a02",
		},
		{
			Filename:         "V3__test_remove_column_from_product.sql",
			ExpectedChecksum: "ad237d5f6002d5dbad359f98e2ef38dc74bb0e4eb838dafe923cfe7229b5024c",
		},
	}

	for _, f := range fileList {
		res, err := CalculateChecksum(getWorkPath() + "/utils/test/db/migration/postgres/" + f.Filename)
		if err != nil {
			t.Errorf("expected nil but got error %v", err)
		}

		if res != f.ExpectedChecksum {
			t.Errorf("expected checksum %s but got %s for file %s", f.ExpectedChecksum, res, f.Filename)
		}
	}
}

func getDatabaseMigrations() []historyModel {
	currentTime := time.Now()
	dbMigrations := []historyModel{
		{
			InstalledRank: 1,
			Version:       "1",
			Description:   "test create table product",
			Type:          "sql",
			Script:        "V1__test_create_table_product.sql",
			Checksum:      "423156a418cd885a0304a6f0cc2ad8e7059e8421ae181800257fa31815e9d197",
			InstalledBy:   "root",
			InstalledOn:   &currentTime,
			ExecutionTime: 17,
			Success:       true,
		},
		{
			InstalledRank: 2,
			Version:       "2",
			Description:   "test alter table product",
			Type:          "sql",
			Script:        "V2__test_alter_table_product.sql",
			Checksum:      "2eaef764b14c8a99535a61a9f4fd4af9428e1905bfa68659e6af2ed2c45d3a02",
			InstalledBy:   "root",
			InstalledOn:   &currentTime,
			ExecutionTime: 17,
			Success:       true,
		},
		{
			InstalledRank: 3,
			Version:       "3",
			Description:   "test remove column from product",
			Type:          "sql",
			Script:        "V3__test_remove_column_from_product.sql",
			Checksum:      "ad237d5f6002d5dbad359f98e2ef38dc74bb0e4eb838dafe923cfe7229b5024c",
			InstalledBy:   "root",
			InstalledOn:   &currentTime,
			ExecutionTime: 17,
			Success:       true,
		},
	}

	return dbMigrations
}

func getLocalMigrations() []localScript {
	localMigrations := []localScript{
		{
			Version:     "1",
			Description: "test create table product",
			Script:      "V1__test_create_table_product.sql",
			Checksum:    "423156a418cd885a0304a6f0cc2ad8e7059e8421ae181800257fa31815e9d197",
		},
		{
			Version:     "2",
			Description: "test alter table product",
			Script:      "V2__test_alter_table_product.sql",
			Checksum:    "2eaef764b14c8a99535a61a9f4fd4af9428e1905bfa68659e6af2ed2c45d3a02",
		},
		{
			Version:     "3",
			Description: "test remove column from product",
			Script:      "V3__test_remove_column_from_product.sql",
			Checksum:    "ad237d5f6002d5dbad359f98e2ef38dc74bb0e4eb838dafe923cfe7229b5024c",
		},
	}

	return localMigrations
}
