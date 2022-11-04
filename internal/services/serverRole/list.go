package serverRole

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

func (l listDataSource) GetSchema(ctx context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}

	for name, attr := range attributes {
		attr.Computed = true
		attrs[name] = attr
	}

	return tfsdk.Schema{
		MarkdownDescription: "Obtains information about all roles defined in the server.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Only used internally in Terraform. Always set to `server_roles`.",
				Type:                types.StringType,
				Computed:            true,
			},
			"roles": {
				MarkdownDescription: "Set of all roles found in the server",
				Computed:            true,

				Attributes: tfsdk.SetNestedAttributes(attrs),
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
