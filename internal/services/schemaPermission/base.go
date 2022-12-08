package schemaPermission

var attrDescriptions = map[string]string{
	"id":                "`<database_id>/<schema_id>/<principal_id>/<permission>`.",
	"schema_id":         "`<database_id>/<schema_id>`. Can be retrieved using `mssql_schema`.",
	"principal_id":      "`<database_id>/<principal_id>`. Can be retrieved using `mssql_database_role`, `mssql_sql_user`, `mssql_azuread_user` or `mssql_azuread_service_principal`.",
	"permission":        "Name of schema SQL permission. For full list of supported permissions, see [docs](https://learn.microsoft.com/en-us/sql/t-sql/statements/grant-schema-permissions-transact-sql?view=azuresqldb-current#remarks)",
	"with_grant_option": "When set to `true`, `principal_id` will be allowed to grant the `permission` to other principals.",
}
