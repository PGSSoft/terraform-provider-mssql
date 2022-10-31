package schemaPermission

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceDataPermission struct {
	Permission      types.String `tfsdk:"permission"`
	WithGrantOption types.Bool   `tfsdk:"with_grant_option"`
}

type listDataSourceData struct {
	Id          types.String                   `tfsdk:"id"`
	SchemaId    types.String                   `tfsdk:"schema_id"`
	PrincipalId types.String                   `tfsdk:"principal_id"`
	Permissions []listDataSourceDataPermission `tfsdk:"permissions"`
}

var _ datasource.DataSourceWithValidation[listDataSourceData] = listDataSource{}

type listDataSource struct{}

func (l listDataSource) GetName() string {
	return "schema_permissions"
}

func (l listDataSource) GetSchema(ctx context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: "Returns all permissions granted in a schema to given principal",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "`<database_id>/<schema_id>/<principal_id>`.",
				Type:                types.StringType,
				Computed:            true,
			},
			"schema_id":    common.ToRequired(attributes["schema_id"]),
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
	schemaId, principalId := l.parseInputs(ctx, req.Config)

	var (
		db     sql.Database
		schema sql.Schema
		perms  sql.SchemaPermissions
	)

	req.
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, schemaId.DbId) }).
		Then(func() { schema = sql.GetSchema(ctx, db, schemaId.ObjectId) }).
		Then(func() { perms = schema.GetPermissions(ctx, principalId.ObjectId) }).
		Then(func() {
			state := req.Config
			state.Id = types.StringValue(fmt.Sprintf("%s/%d", schemaId, principalId.ObjectId))

			for _, perm := range perms {
				state.Permissions = append(state.Permissions, listDataSourceDataPermission{
					Permission:      types.StringValue(perm.Name),
					WithGrantOption: types.BoolValue(perm.WithGrantOption),
				})
			}

			resp.SetState(state)
		})
}

func (l listDataSource) Validate(ctx context.Context, req datasource.ValidateRequest[listDataSourceData], _ *datasource.ValidateResponse[listDataSourceData]) {
	schemaId, principalId := l.parseInputs(ctx, req.Config)

	req.Then(func() {
		if schemaId.DbId != principalId.DbId {
			err := fmt.Errorf("schema_id points to DB with ID %d while principal_id points to DB with ID %d", schemaId.DbId, principalId.DbId)
			utils.AddError(ctx, "Schema and Principal must belong to the same DB", err)
		}
	})
}

func (l listDataSource) parseInputs(ctx context.Context, data listDataSourceData) (common.DbObjectId[sql.SchemaId], common.DbObjectId[sql.GenericDatabasePrincipalId]) {
	return common.ParseDbObjectId[sql.SchemaId](ctx, data.SchemaId.ValueString()), common.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, data.PrincipalId.ValueString())
}
