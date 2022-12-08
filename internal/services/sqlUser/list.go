package sqlUser

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
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

func (l *listDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about all SQL users found in a database"
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "ID of the resource, equals to database ID",
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common.AttributeDescriptions["database_id"] + " Defaults to ID of `master`.",
			Optional:            true,
		},
		"users": schema.SetNestedAttribute{
			Description: "Set of SQL user objects",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: attrDescriptions["id"],
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: attrDescriptions["name"],
						Computed:            true,
					},
					"database_id": schema.StringAttribute{
						MarkdownDescription: common.AttributeDescriptions["database_id"],
						Computed:            true,
					},
					"login_id": schema.StringAttribute{
						MarkdownDescription: attrDescriptions["login_id"],
						Computed:            true,
					},
				},
			},
		},
	}
}

func (l *listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	var db sql.Database
	var dbId sql.DatabaseId

	req.
		Then(func() { db = common.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString()) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() {
			state := listDataSourceData{
				DatabaseId: types.StringValue(fmt.Sprint(dbId)),
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
