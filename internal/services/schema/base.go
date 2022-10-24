package schema

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var attributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<schema_id>`. Schema ID can be retrieved using `SELECT SCHEMA_ID('<schema_name>')`.",
		Type:                types.StringType,
	},
	"database_id": common.DatabaseIdAttribute,
	"name": {
		MarkdownDescription: "Schema name.",
		Type:                types.StringType,
		Validators:          validators.SchemaNameValidators,
	},
	"owner_id": {
		MarkdownDescription: "ID of database role or user owning this schema. Can be retrieved using `mssql_database_role`, `mssql_sql_user`, `mssql_azuread_user` or `mssql_azuread_service_principal`",
		Type:                types.StringType,
	},
}

type resourceData struct {
	Id         types.String `tfsdk:"id"`
	DatabaseId types.String `tfsdk:"database_id"`
	Name       types.String `tfsdk:"name"`
	OwnerId    types.String `tfsdk:"owner_id"`
}

func (d resourceData) withSchemaData(ctx context.Context, schema sql.Schema) resourceData {
	dbId := schema.GetDb(ctx).GetId(ctx)

	return resourceData{
		Id:         types.String{Value: common.DbObjectId[sql.SchemaId]{DbId: dbId, ObjectId: schema.GetId(ctx)}.String()},
		Name:       d.Name,
		DatabaseId: types.String{Value: fmt.Sprint(dbId)},
		OwnerId:    types.String{Value: common.DbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: schema.GetOwnerId(ctx)}.String()},
	}
}
