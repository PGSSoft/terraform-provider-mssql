package serverPermission

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/attrs"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (l listDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Returns all permissions grated to given principal"
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			CustomType:          attrs.NumericIdType[sql.GenericServerPrincipalId](),
			MarkdownDescription: "Equals to `principal_id`.",
			Computed:            true,
		},
		"principal_id": schema.StringAttribute{
			CustomType:          attrs.NumericIdType[sql.GenericServerPrincipalId](),
			MarkdownDescription: attrDescriptions["principal_id"],
			Required:            true,
		},
		"permissions": schema.SetNestedAttribute{
			MarkdownDescription: "Set of permissions granted to the principal",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"permission": schema.StringAttribute{
						MarkdownDescription: attrDescriptions["permission"],
						Computed:            true,
					},
					"with_grant_option": schema.BoolAttribute{
						MarkdownDescription: attrDescriptions["with_grant_option"],
						Computed:            true,
					},
				},
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
