package serverRole

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceData struct {
	Id    types.String   `tfsdk:"id"`
	Roles []resourceData `tfsdk:"roles"`
}

type listDataSource struct{}

func (l listDataSource) GetName() string {
	return "server_roles"
}

func (l listDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about all roles defined in the server."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Only used internally in Terraform. Always set to `server_roles`.",
			Computed:            true,
		},
		"roles": schema.SetNestedAttribute{
			MarkdownDescription: "Set of all roles found in the server",
			Computed:            true,
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
					"owner_id": schema.StringAttribute{
						MarkdownDescription: attrDescriptions["owner_id"],
						Computed:            true,
					},
				},
			},
		},
	}
}

func (l listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	roles := sql.GetServerRoles(ctx, req.Conn)

	req.
		Then(func() {
			data := listDataSourceData{
				Id:    types.StringValue("server_roles"),
				Roles: []resourceData{},
			}

			for id, role := range roles {
				roleData := resourceData{Id: types.StringValue(fmt.Sprint(id))}.withSettings(role.GetSettings(ctx))
				data.Roles = append(data.Roles, roleData)
			}

			resp.SetState(data)
		})
}
