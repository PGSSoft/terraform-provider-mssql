package provider

import (
	"context"
	"fmt"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ datasource.DataSourceWithConfigure = &databaseRoleList{}
)

type databaseRoleListData struct {
	Id         types.String               `tfsdk:"id"`
	DatabaseId types.String               `tfsdk:"database_id"`
	Roles      []databaseRoleResourceData `tfsdk:"roles"`
}

type databaseRoleList struct {
	Resource
}

func (p mssqlProvider) NewDatabaseRoleListDataSource() func() datasource.DataSource {
	return func() datasource.DataSource {
		return databaseRoleList{}
	}
}

func (s *databaseRoleList) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	s.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (s databaseRoleList) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "mssql_database_roles"
}

func (d databaseRoleList) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range databaseRoleAttributes {
		attr.Computed = true
		attrs[n] = attr
	}

	return tfsdk.Schema{
		Description: "Obtains information about all roles defined in a database.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "ID of the resource, equals to database ID",
			},
			"database_id": databaseIdResourceAttribute,
			"roles": {
				Description: "Set of database role objects",
				Attributes:  tfsdk.SetNestedAttributes(attrs),
				Computed:    true,
			},
		},
	}, nil
}

func (d databaseRoleList) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		data  databaseRoleListData
		db    sql.Database
		dbId  sql.DatabaseId
		roles map[sql.DatabaseRoleId]sql.DatabaseRole
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleListData](ctx, request.Config) }).
		Then(func() { db = getResourceDb(ctx, d.Db, data.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() { roles = sql.GetDatabaseRoles(ctx, db) }).
		Then(func() {
			data.DatabaseId = types.String{Value: fmt.Sprint(dbId)}
			data.Id = data.DatabaseId

			for _, role := range roles {
				data.Roles = append(data.Roles, databaseRoleResourceData{}.withRoleData(ctx, role))
			}
		}).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}
