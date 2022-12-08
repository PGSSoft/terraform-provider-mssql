package schema

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var attrDescriptions = map[string]string{
	"id":       "`<database_id>/<schema_id>`. Schema ID can be retrieved using `SELECT SCHEMA_ID('<schema_name>')`.",
	"name":     "Schema name.",
	"owner_id": "ID of database role or user owning this schema. Can be retrieved using `mssql_database_role`, `mssql_sql_user`, `mssql_azuread_user` or `mssql_azuread_service_principal`",
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
		Id:         types.StringValue(common.DbObjectId[sql.SchemaId]{DbId: dbId, ObjectId: schema.GetId(ctx)}.String()),
		Name:       types.StringValue(schema.GetName(ctx)),
		DatabaseId: types.StringValue(fmt.Sprint(dbId)),
		OwnerId:    types.StringValue(common.DbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: schema.GetOwnerId(ctx)}.String()),
	}
}
