package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type databaseRoleListData struct {
	Id         types.String               `tfsdk:"id"`
	DatabaseId types.String               `tfsdk:"database_id"`
	Roles      []databaseRoleResourceData `tfsdk:"roles"`
}

type databaseRoleList struct {
	BaseDataSource
}

func (p mssqlProvider) NewDatabaseRoleListDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[databaseRoleListData](&databaseRoleList{})
	}
}

func (l *databaseRoleList) GetName() string {
	return "database_roles"
}

func (l *databaseRoleList) GetSchema(context.Context) tfsdk.Schema {
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
	}
}

func (l *databaseRoleList) Read(ctx context.Context, req datasource.ReadRequest[databaseRoleListData], resp *datasource.ReadResponse[databaseRoleListData]) {
	var (
		db    sql.Database
		dbId  sql.DatabaseId
		roles map[sql.DatabaseRoleId]sql.DatabaseRole
	)

	req.
		Then(func() { db = getResourceDb(ctx, l.conn, req.Config.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() { roles = sql.GetDatabaseRoles(ctx, db) }).
		Then(func() {
			state := databaseRoleListData{
				DatabaseId: types.String{Value: fmt.Sprint(dbId)},
			}
			state.Id = state.DatabaseId

			for _, role := range roles {
				state.Roles = append(state.Roles, databaseRoleResourceData{}.withRoleData(ctx, role))
			}

			resp.SetState(state)
		})
}
