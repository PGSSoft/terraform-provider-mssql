package schemaPermission

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (l listDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Returns all permissions granted in a schema to given principal"
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "`<database_id>/<schema_id>/<principal_id>`.",
			Computed:            true,
		},
		"schema_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["schema_id"],
			Required:            true,
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
