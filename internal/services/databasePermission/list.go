package databasePermission

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceDataPermission struct {
	Permission      types.String `tfsdk:"permission"`
	WithGrantOption types.Bool   `tfsdk:"with_grant_option"`
}

type listDataSourceData struct {
	Id          types.String                   `tfsdk:"id"`
	PrincipalId types.String                   `tfsdk:"principal_id"`
	Permissions []listDataSourceDataPermission `tfsdk:"permissions"`
}

type listDataSource struct{}

func (l listDataSource) GetName() string {
	return "database_permissions"
}

func (l listDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Returns all permissions granted in a DB to given principal"
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "`<database_id>/<principal_id>`.",
			Computed:            true,
		},
		"principal_id": schema.StringAttribute{
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
	var db sql.Database
	var perms sql.DatabasePermissions

	principalId := common.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Config.PrincipalId.ValueString())

	req.
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, principalId.DbId) }).
		Then(func() { perms = db.GetPermissions(ctx, principalId.ObjectId) }).
		Then(func() {
			req.Config.Id = req.Config.PrincipalId

			for _, perm := range perms {
				req.Config.Permissions = append(req.Config.Permissions, listDataSourceDataPermission{
					Permission:      types.StringValue(perm.Name),
					WithGrantOption: types.BoolValue(perm.WithGrantOption),
				})
			}

			resp.SetState(req.Config)
		})
}
