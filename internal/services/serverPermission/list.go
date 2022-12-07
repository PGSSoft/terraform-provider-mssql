package serverPermission

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/attrs"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceDataPermission struct {
	Permission      types.String `tfsdk:"permission"`
	WithGrantOption types.Bool   `tfsdk:"with_grant_option"`
}

type listDataSourceData struct {
	Id          attrs.NumericId[sql.GenericServerPrincipalId] `tfsdk:"id"`
	PrincipalId attrs.NumericId[sql.GenericServerPrincipalId] `tfsdk:"principal_id"`
	Permissions []listDataSourceDataPermission                `tfsdk:"permissions"`
}

type listDataSource struct{}

func (l listDataSource) GetName() string {
	return "server_permissions"
}

func (l listDataSource) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: "Returns all permissions grated to given principal",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Equals to `principal_id`.",
				Type:                attrs.NumericIdType[sql.GenericServerPrincipalId](),
				Computed:            true,
			},
			"principal_id": common.ToRequired(attributes["principal_id"]),
			"permissions": {
				MarkdownDescription: "Set of permissions granted to the principal",
				Computed:            true,

				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"permission": func() tfsdk.Attribute {
						attr := attributes["permission"]
						attr.Computed = true
						return attr
					}(),
					"with_grant_option": func() tfsdk.Attribute {
						attr := attributes["with_grant_option"]
						attr.Computed = true
						return attr
					}(),
				}),
			},
		},
	}
}

func (l listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	principalId := req.Config.PrincipalId.Id(ctx)
	var perms sql.ServerPermissions

	req.
		Then(func() { perms = req.Conn.GetPermissions(ctx, principalId) }).
		Then(func() {
			req.Config.Id = req.Config.PrincipalId
			req.Config.Permissions = []listDataSourceDataPermission{}
			for _, perm := range perms {
				req.Config.Permissions = append(req.Config.Permissions, listDataSourceDataPermission{
					Permission:      types.StringValue(perm.Name),
					WithGrantOption: types.BoolValue(perm.WithGrantOption),
				})
			}

			resp.SetState(req.Config)
		})
}
