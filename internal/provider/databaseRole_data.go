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
	_ tfsdk.DataSourceType = DatabaseRoleDataSourceType{}
	_ tfsdk.DataSource     = databaseRoleData{}
)

type DatabaseRoleDataSourceType struct{}

func (d DatabaseRoleDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range databaseRoleAttributes {
		if n == "database_id" {
			attr = databaseIdResourceAttribute
			attr.Optional = true
		}
		
		attr.Required = n == "name"
		attr.Computed = n != "name"

		attrs[n] = attr
	}

	return tfsdk.Schema{
		Description: "Obtains information about single database role.",
		Attributes:  attrs,
	}, nil
}

func (d DatabaseRoleDataSourceType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) databaseRoleData {
		return databaseRoleData{Resource: base}
	})
}

type databaseRoleData struct {
	Resource
}

func (d databaseRoleData) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	var (
		data databaseRoleResourceData
		db   sql.Database
		role sql.DatabaseRole
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleResourceData](ctx, request.Config) }).
		Then(func() { db = getResourceDb(ctx, d.Db, data.DatabaseId.Value) }).
		Then(func() { role = sql.GetDatabaseRoleByName(ctx, db, data.Name.Value) }).
		Then(func() { data = data.withRoleData(ctx, role) }).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}
