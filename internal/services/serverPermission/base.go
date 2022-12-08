package serverPermission

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/attrs"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var attributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<principal_id>/<permission>`",
		Type:                attrs.PermissionIdType[sql.GenericServerPrincipalId](),
	},
	"principal_id": {
		MarkdownDescription: "ID of the principal who will be granted `permission`. Can be retrieved using `mssql_server_role` or `mssql_sql_login`.",
		Type:                attrs.NumericIdType[sql.GenericServerPrincipalId](),
	},
	"permission": {
		MarkdownDescription: "Name of server-level SQL permission. For full list of supported permissions see [docs](https://learn.microsoft.com/en-us/sql/t-sql/statements/grant-server-permissions-transact-sql?view=azuresqldb-current#remarks)",
		Type:                types.StringType,
	},
	"with_grant_option": {
		MarkdownDescription: "When set to `true`, `principal_id` will be allowed to grant the `permission` to other principals.",
		Type:                types.BoolType,
	},
}

var attrDescriptions = map[string]string{
	"id":                "`<principal_id>/<permission>`",
	"principal_id":      "ID of the principal who will be granted `permission`. Can be retrieved using `mssql_server_role` or `mssql_sql_login`.",
	"permission":        "Name of server-level SQL permission. For full list of supported permissions see [docs](https://learn.microsoft.com/en-us/sql/t-sql/statements/grant-server-permissions-transact-sql?view=azuresqldb-current#remarks)",
	"with_grant_option": "When set to `true`, `principal_id` will be allowed to grant the `permission` to other principals.",
}
