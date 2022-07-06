package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var databaseRoleMemberAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<role_id>/<member_id>`. Role and member IDs can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<name>')`",
		Type:                types.StringType,
	},
	"role_id": {
		MarkdownDescription: "`<database_id>/<role_id>`",
		Type:                types.StringType,
	},
	"member_id": {
		MarkdownDescription: "Can be either user or role ID in format `<database_id>/<member_id>`. Can be retrieved using `mssql_sql_user` or `mssql_database_member`.",
		Type:                types.StringType,
	},
}

type databaseRoleMemberResourceData struct {
	Id       types.String `tfsdk:"id"`
	RoleId   types.String `tfsdk:"role_id"`
	MemberId types.String `tfsdk:"member_id"`
}
