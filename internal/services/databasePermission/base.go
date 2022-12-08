package databasePermission

var attrDescriptions = map[string]string{
	"id":                "`<database_id>/<principal_id>/<permission>`.",
	"principal_id":      "`<database_id>/<principal_id>`. Can be retrieved using `mssql_database_role`, `mssql_sql_user`, `mssql_azuread_user` or `mssql_azuread_service_principal`.",
	"permission":        "Name of database-level SQL permission. For full list of supported permissions, see [docs](https://learn.microsoft.com/en-us/sql/t-sql/statements/grant-database-permissions-transact-sql?view=azuresqldb-current#remarks)",
	"with_grant_option": "When set to `true`, `principal_id` will be allowed to grant the `permission` to other principals.",
}
