package provider

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.DataSourceType = SqlUserDataSourceType{}
	_ tfsdk.DataSource     = sqlUserData{}
)

type SqlUserDataSourceType struct{}

func (s SqlUserDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	a := map[string]tfsdk.Attribute{}
	for n, attribute := range sqlUserAttributes {
		attribute.Required = n == "name"
		attribute.Optional = n == "database_id"
		attribute.Computed = n != "name"
		a[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about single SQL database user.",
		Attributes:  a,
	}, nil
}

func (s SqlUserDataSourceType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) sqlUserData {
		return sqlUserData{sqlUserResourceBase: sqlUserResourceBase{Resource: base}}
	})
}

type sqlUserData struct {
	sqlUserResourceBase
}

func (s sqlUserData) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlUserResourceData](ctx, request.Config)
	if utils.HasError(ctx) {
		return
	}

	db := getResourceDb(ctx, s.Db, data.DatabaseId.Value)
	if utils.HasError(ctx) {
		return
	}

	user := sql.GetUserByName(ctx, db, data.Name.Value)
	if utils.HasError(ctx) {
		return
	}

	data = data.withIds(db.GetId(ctx), user.GetId(ctx)).withSettings(user.GetSettings(ctx))
	if utils.HasError(ctx) {
		return
	}

	utils.SetData(ctx, &response.State, data)
}
