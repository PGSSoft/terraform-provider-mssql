package databaseRole

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "database_role"
}

func (d *dataSource) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{
		"members": {
			Description: "Set of role members",
			Attributes:  tfsdk.SetNestedAttributes(roleMemberAttributes),
			Computed:    true,
		},
	}

	for n, attr := range roleAttributes {
		if n == "database_id" {
			attr = common2.DatabaseIdResourceAttribute
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

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[dataSourceData], resp *datasource.ReadResponse[dataSourceData]) {
	var (
		db   sql.Database
		role sql.DatabaseRole
	)

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.Value) }).
		Then(func() { role = sql.GetDatabaseRoleByName(ctx, db, req.Config.Name.Value) }).
		Then(func() { resp.SetState(req.Config.withRoleData(ctx, role)) })
}
