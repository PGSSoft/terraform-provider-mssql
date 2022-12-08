package sqlLogin

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "sql_login"
}

func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about single SQL login."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["id"],
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["name"],
			Required:            true,
		},
		"must_change_password": schema.BoolAttribute{
			MarkdownDescription: attrDescriptions["must_change_password"],
			Computed:            true,
		},
		"default_database_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["default_database_id"],
			Computed:            true,
		},
		"default_language": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["default_language"],
			Computed:            true,
		},
		"check_password_expiration": schema.BoolAttribute{
			MarkdownDescription: attrDescriptions["check_password_expiration"],
			Computed:            true,
		},
		"check_password_policy": schema.BoolAttribute{
			MarkdownDescription: attrDescriptions["check_password_policy"],
			Computed:            true,
		},
		"principal_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["principal_id"],
			Computed:            true,
		},
	}
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[dataSourceData], resp *datasource.ReadResponse[dataSourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() {
			login = sql.GetSqlLoginByName(ctx, req.Conn, req.Config.Name.ValueString())

			if login == nil || !login.Exists(ctx) {
				utils.AddError(ctx, "Login does not exist", fmt.Errorf("could not find SQL Login '%s'", req.Config.Name.ValueString()))
			}
		}).
		Then(func() {
			state := req.Config.withSettings(login.GetSettings(ctx))
			state.Id = types.StringValue(fmt.Sprint(login.GetId(ctx)))

			resp.SetState(state)
		})
}
