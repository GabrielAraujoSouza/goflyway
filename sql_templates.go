package goflyway

const createTablePostgres = `
	CREATE TABLE IF NOT EXISTS "[tableName]" (
		installed_rank BIGINT NOT NULL,
		"version" VARCHAR(255),
		description VARCHAR(255),
		"type" VARCHAR(50),
		"script" VARCHAR(255),
		checksum VARCHAR(255),
		installed_by VARCHAR(255),
		installed_on TIMESTAMP,
		execution_time BIGINT,
		success BOOLEAN,

		CONSTRAINT pk_goflyway_sch_hist PRIMARY KEY (installed_rank)
	)
`

const createTableMysql = "CREATE TABLE IF NOT EXISTS `[tableName]` (" +
	" installed_rank BIGINT NOT NULL, " +
	" `version` VARCHAR(255), " +
	" description VARCHAR(255), " +
	" `type` VARCHAR(50), " +
	" `script` VARCHAR(255), " +
	" checksum VARCHAR(255), " +
	" installed_by VARCHAR(255), " +
	" installed_on TIMESTAMP, " +
	" execution_time BIGINT, " +
	" success BOOLEAN, " +
	"	CONSTRAINT pk_goflyway_sch_hist PRIMARY KEY (installed_rank) " +
	" ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE utf8_bin; "

const selectTablePostgres = `
	SELECT installed_rank, "version", description, "type", "script", 
	  	   checksum, installed_by, installed_on, execution_time, success
	FROM "[tableName]" ORDER BY "version"
`
const selectTableMysql = "SELECT installed_rank, `version`, description, `type`, `script`," +
	" checksum, installed_by, installed_on, execution_time, success" +
	" FROM `[tableName]` ORDER BY version"

const insertPostgres = `
	INSERT INTO "[tableName]"
	(installed_rank, "version", description, "type", script, checksum, installed_by, installed_on, execution_time, success)
	VALUES($1, $2, $3, $4, $5, $6, current_user, current_timestamp, $7, true);
`

const insertMysql = "INSERT INTO `[tableName]` " +
	"(installed_rank, `version`, description, `type`, `script`, checksum, installed_by, installed_on, execution_time, success)" +
	" VALUES(?, ?, ?, ?, ?, ?, current_user, current_timestamp, ?, true)"
