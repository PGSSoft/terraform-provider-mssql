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

func (p mssqlProvider) NewSqlUserListDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[sqlUserListData](&sqlUserList{})
	}
}

func (l *sqlUserList) GetName() string {
	return "sql_users"
}

type sqlUserListData struct {
	Id         types.String          `tfsdk:"id"`
	DatabaseId types.String          `tfsdk:"database_id"`
	Users      []sqlUserResourceData `tfsdk:"users"`
}

type sqlUserList struct {
	BaseDataSource
}

func (l *sqlUserList) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range sqlUserAttributes {
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
				attr := sqlUserAttributes["database_id"]
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

func (l *sqlUserList) Read(ctx context.Context, req datasource.ReadRequest[sqlUserListData], resp *datasource.ReadResponse[sqlUserListData]) {
	var db sql.Database
	var dbId sql.DatabaseId

	req.
		Then(func() { db = getResourceDb(ctx, l.conn, req.Config.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() {
			state := sqlUserListData{
				DatabaseId: types.String{Value: fmt.Sprint(dbId)},
			}
			state.Id = state.DatabaseId

			for id, user := range sql.GetUsers(ctx, db) {
				s := user.GetSettings(ctx)

				if s.Type == sql.USER_TYPE_SQL {
					state.Users = append(state.Users, sqlUserResourceData{}.withIds(dbId, id).withSettings(s))
				}
			}

			resp.SetState(state)
		})
}
