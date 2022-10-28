package databasePermission

import (
	"context"
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
	Id          types.String                   `tfsdk:"id"`
	PrincipalId types.String                   `tfsdk:"principal_id"`
	Permissions []listDataSourceDataPermission `tfsdk:"permissions"`
}

type listDataSource struct{}

func (l listDataSource) GetName() string {
	return "database_permissions"
}

func (l listDataSource) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: "Returns all permissions granted in a DB to given principal",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "`<database_id>/<principal_id>`.",
				Type:                types.StringType,
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
