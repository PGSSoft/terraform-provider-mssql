package schemaPermission

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var attributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<schema_id>/<principal_id>/<permission>`.",
		Type:                types.StringType,
	},
	"schema_id": {
		MarkdownDescription: "`<database_id>/<schema_id`. Can be retrieved using `mssql_schema`.",
		Type:                types.StringType,
	},
	"principal_id": {
		MarkdownDescription: "`<database_id>/<principal_id>`. Can be retrieved using `mssql_database_role`, `mssql_sql_user`, `mssql_azuread_user` or `mssql_azuread_service_principal`.",
		Type:                types.StringType,
	},
	"permission": {
		MarkdownDescription: "Name of schema SQL permission. For full list of supported permissions, see [docs](https://learn.microsoft.com/en-us/sql/t-sql/statements/grant-schema-permissions-transact-sql?view=azuresqldb-current#remarks)",
		Type:                types.StringType,
	},
	"with_grant_option": {
		MarkdownDescription: "When set to `true`, `principal_id` will be allowed to grant the `permission` to other principals.",
		Type:                types.BoolType,
	},
}
