package sqlUser

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
	Users      []resourceData `tfsdk:"users"`
}

type listDataSource struct{}

func (l *listDataSource) GetName() string {
	return "sql_users"
}

func (l *listDataSource) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range attributes {
		attr.Computed = true
		attrs[n] = attr
	}

	return tfsdk.Schema{
		Description: "Obtains information about all SQL users found in a database",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "ID of the resource, equals to database ID",
			},
			"database_id": func() tfsdk.Attribute {
				attr := attributes["database_id"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to ID of `master`."
				return attr
			}(),
			"users": {
				Description: "Set of SQL user objects",
				Attributes:  tfsdk.SetNestedAttributes(attrs),
				Computed:    true,
			},
		},
	}
}

func (l *listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	var db sql.Database
	var dbId sql.DatabaseId

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() {
			state := listDataSourceData{
				DatabaseId: types.String{Value: fmt.Sprint(dbId)},
			}
			state.Id = state.DatabaseId

			for id, user := range sql.GetUsers(ctx, db) {
				s := user.GetSettings(ctx)

				if s.Type == sql.USER_TYPE_SQL {
					state.Users = append(state.Users, resourceData{}.withIds(dbId, id).withSettings(s))
				}
			}

			resp.SetState(state)
		})
}
