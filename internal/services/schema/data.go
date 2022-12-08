package schema

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSourceWithValidation[resourceData] = dataSource{}

type dataSource struct{}

func (d dataSource) GetName() string {
	return "schema"
}

func (d dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Retrieves information about DB schema."

	const idNameRemark = " Either `id` or `name` must be provided."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["id"] + idNameRemark,
			Computed:            true,
			Optional:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["name"] + idNameRemark,
			Computed:            true,
			Optional:            true,
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common.AttributeDescriptions["database_id"],
			Optional:            true,
			Computed:            true,
		},
		"owner_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["owner_id"],
			Computed:            true,
		},
	}
}

func (d dataSource) Read(ctx context.Context, req datasource.ReadRequest[resourceData], resp *datasource.ReadResponse[resourceData]) {
	var (
		db     sql.Database
		schema sql.Schema
	)

	schemaId := common.ParseDbObjectId[sql.SchemaId](ctx, req.Config.Id.ValueString())

	req.
		Then(func() {
			if schemaId.IsEmpty {
				db = common.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString())
			} else {
				db = sql.GetDatabase(ctx, req.Conn, schemaId.DbId)
			}
		}).
		Then(func() {
			if schemaId.IsEmpty {
				schema = sql.GetSchemaByName(ctx, db, req.Config.Name.ValueString())
			} else {
				schema = sql.GetSchema(ctx, db, schemaId.ObjectId)
			}
		}).
		Then(func() {
			resp.SetState(req.Config.withSchemaData(ctx, schema))
		})
}

func (d dataSource) Validate(ctx context.Context, req datasource.ValidateRequest[resourceData], _ *datasource.ValidateResponse[resourceData]) {
	if !common.IsAttrSet(req.Config.Id) && !common.IsAttrSet(req.Config.Name) {
		utils.AddError(ctx, "One of id or name must be provided", errors.New("both id and name attributes are unknown"))
	}
}
