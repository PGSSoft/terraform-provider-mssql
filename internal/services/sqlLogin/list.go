package sqlLogin

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceData struct {
	Id     types.String     `tfsdk:"id"`
	Logins []dataSourceData `tfsdk:"logins"`
}

type listDataSource struct{}

func (l *listDataSource) GetName() string {
	return "sql_logins"
}

func (l *listDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about all SQL logins found in SQL Server instance."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "ID of the resource used only internally by the provider.",
		},
		"logins": schema.SetNestedAttribute{
			Description: "Set of SQL login objects",
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
					}},
			},
		},
	}
}

func (l *listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	var logins map[sql.LoginId]sql.SqlLogin

	req.
		Then(func() { logins = sql.GetSqlLogins(ctx, req.Conn) }).
		Then(func() {
			result := listDataSourceData{
				Id: types.StringValue(""),
			}

			for id, login := range logins {
				s := login.GetSettings(ctx)
				r := dataSourceData{Id: types.StringValue(fmt.Sprint(id))}
				result.Logins = append(result.Logins, r.withSettings(s))
			}

			resp.SetState(result)
		})
}
