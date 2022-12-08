package sqlUser

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "sql_user"
}

func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about single SQL database user."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["id"],
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["name"],
			Required:            true,
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common2.AttributeDescriptions["database_id"],
			Optional:            true,
			Computed:            true,
		},
		"login_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["login_id"],
			Computed:            true,
		},
	}
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[resourceData], resp *datasource.ReadResponse[resourceData]) {
	var db sql.Database
	var user sql.User

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString()) }).
		Then(func() { user = sql.GetUserByName(ctx, db, req.Config.Name.ValueString()) }).
		Then(func() {
			state := req.Config.withIds(db.GetId(ctx), user.GetId(ctx))
			resp.SetState(state.withSettings(user.GetSettings(ctx)))
		})
}
