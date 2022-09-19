package provider

import (
	"context"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ datasource.DataSourceWithConfigure = &databaseRoleData{}
)

type databaseRoleData struct {
	Resource
}

func (p mssqlProvider) NewDatabaseRoleDataSource() func() datasource.DataSource {
	return func() datasource.DataSource {
		return databaseRoleData{}
	}
}

func (s *databaseRoleData) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	s.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (s databaseRoleData) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "mssql_database_role"
}

func (d databaseRoleData) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
	}, nil
}

func (d databaseRoleData) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		data databaseRoleDataResourceData
		db   sql.Database
		role sql.DatabaseRole
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleDataResourceData](ctx, request.Config) }).
		Then(func() { db = getResourceDb(ctx, d.Db, data.DatabaseId.Value) }).
		Then(func() { role = sql.GetDatabaseRoleByName(ctx, db, data.Name.Value) }).
		Then(func() { data = data.withRoleData(ctx, role) }).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}
