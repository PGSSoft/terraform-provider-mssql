package databasePermission

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var attributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<principal_id>/<permission>`.",
		Type:                types.StringType,
	},
	"principal_id": {
		MarkdownDescription: "`<database_id>/<principal_id>`. Can be retrieved using `mssql_database_role`, `mssql_sql_user`, `mssql_azuread_user` or `mssql_azuread_service_principal`.",
		Type:                types.StringType,
	},
	"permission": {
		MarkdownDescription: "Name of database-level SQL permission. For full list of supported permissions, see [docs](https://learn.microsoft.com/en-us/sql/t-sql/statements/grant-database-permissions-transact-sql?view=azuresqldb-current#remarks)",
		Type:                types.StringType,
	},
	"with_grant_option": {
		MarkdownDescription: "When set to `true`, `principal_id` will be allowed to grant the `permission` to other principals.",
		Type:                types.BoolType,
	},
}
