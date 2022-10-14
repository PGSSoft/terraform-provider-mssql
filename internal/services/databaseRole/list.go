package databaseRole

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceData struct {
	Id         types.String   `tfsdk:"id"`
	DatabaseId types.String   `tfsdk:"database_id"`
	Roles      []resourceData `tfsdk:"roles"`
}

type listDataSource struct{}

func (l *listDataSource) GetName() string {
	return "database_roles"
}

func (l *listDataSource) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range roleAttributes {
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
			"database_id": common2.DatabaseIdResourceAttribute,
			"roles": {
				Description: "Set of database role objects",
				Attributes:  tfsdk.SetNestedAttributes(attrs),
				Computed:    true,
			},
		},
	}
}

func (l *listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	var (
		db    sql.Database
		dbId  sql.DatabaseId
		roles map[sql.DatabaseRoleId]sql.DatabaseRole
	)

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() { roles = sql.GetDatabaseRoles(ctx, db) }).
		Then(func() {
			state := listDataSourceData{
				DatabaseId: types.String{Value: fmt.Sprint(dbId)},
			}
			state.Id = state.DatabaseId

			for _, role := range roles {
				state.Roles = append(state.Roles, resourceData{}.withRoleData(ctx, role))
			}

			resp.SetState(state)
		})
}
