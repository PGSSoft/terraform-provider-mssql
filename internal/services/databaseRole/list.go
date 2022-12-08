package databaseRole

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
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

func (l *listDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about all roles defined in a database."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "ID of the resource, equals to database ID",
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common2.AttributeDescriptions["database_id"] + " Defaults to ID of `master`.",
			Optional:            true,
		},
		"roles": schema.SetNestedAttribute{
			Description: "Set of database role objects",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: roleAttributeDescriptions["id"],
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: roleAttributeDescriptions["name"],
						Computed:            true,
					},
					"database_id": schema.StringAttribute{
						MarkdownDescription: common2.AttributeDescriptions["database_id"],
						Computed:            true,
					},
					"owner_id": schema.StringAttribute{
						MarkdownDescription: roleAttributeDescriptions["owner_id"],
						Computed:            true,
					},
				},
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
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString()) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() { roles = sql.GetDatabaseRoles(ctx, db) }).
		Then(func() {
			state := listDataSourceData{
				DatabaseId: types.StringValue(fmt.Sprint(dbId)),
			}
			state.Id = state.DatabaseId

			for _, role := range roles {
				state.Roles = append(state.Roles, resourceData{}.withRoleData(ctx, role))
			}

			resp.SetState(state)
		})
}
