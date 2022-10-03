package provider

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type databaseRoleData struct {
	BaseDataSource
}

func (p mssqlProvider) NewDatabaseRoleDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[databaseRoleDataResourceData](&databaseRoleData{})
	}
}

func (d *databaseRoleData) GetName() string {
	return "database_role"
}

func (d *databaseRoleData) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{
		"members": {
			Description: "Set of role members",
			Attributes:  tfsdk.SetNestedAttributes(databaseRoleMemberSetAttributes),
			Computed:    true,
		},
	}

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
	}
}

func (d *databaseRoleData) Read(ctx context.Context, req datasource.ReadRequest[databaseRoleDataResourceData], resp *datasource.ReadResponse[databaseRoleDataResourceData]) {
	var (
		db   sql.Database
		role sql.DatabaseRole
	)

	req.
		Then(func() { db = getResourceDb(ctx, d.conn, req.Config.DatabaseId.Value) }).
		Then(func() { role = sql.GetDatabaseRoleByName(ctx, db, req.Config.Name.Value) }).
		Then(func() { resp.SetState(req.Config.withRoleData(ctx, role)) })
}
