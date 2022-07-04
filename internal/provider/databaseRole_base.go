package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var databaseRoleAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<role_id>`. Role ID can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<role_name>')`",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: fmt.Sprintf("Role name. %s and cannot be longer than 128 chars.", regularIdentifiersDoc),
		Type:                types.StringType,
		Validators:          validators.UserNameValidators,
	},
	"database_id": databaseIdAttribute,
	"owner_id": {
		MarkdownDescription: "ID of another database role or user owning this role. Can be retrieved using `mssql_database_role` or `mssql_sql_user`.",
		Type:                types.StringType,
	},
}

type databaseRoleResourceData struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseId types.String `tfsdk:"database_id"`
	OwnerId    types.String `tfsdk:"owner_id"`
}

func (d databaseRoleResourceData) withRoleData(ctx context.Context, role sql.DatabaseRole) databaseRoleResourceData {
	dbId := role.GetDb(ctx).GetId(ctx)

	return databaseRoleResourceData{
		Id:         types.String{Value: DbObjectId[sql.DatabaseRoleId]{DbId: dbId, ObjectId: role.GetId(ctx)}.String()},
		Name:       types.String{Value: role.GetName(ctx)},
		DatabaseId: types.String{Value: fmt.Sprint(dbId)},
		OwnerId:    types.String{Value: DbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: role.GetOwnerId(ctx)}.String()},
	}
}
