package common

const RegularIdentifiersDoc = "Must follow [Regular Identifiers rules](https://docs.microsoft.com/en-us/sql/relational-databases/databases/database-identifiers#rules-for-regular-identifiers)"

var AttributeDescriptions = map[string]string{
	"database_id": "ID of database. Can be retrieved using `mssql_database` or `SELECT DB_ID('<db_name>')`.",
}
