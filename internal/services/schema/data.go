package schema

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type dataSource struct{}

func (d dataSource) GetName() string {
	return "schema"
}

func (d dataSource) GetSchema(ctx context.Context) tfsdk.Schema {
	const idNameRemark = " Either `id` or `name` must be provided."

	return tfsdk.Schema{
		MarkdownDescription: "Retrieves information about DB schema.",
		Attributes: map[string]tfsdk.Attribute{
			"id": func() tfsdk.Attribute {
				attr := attributes["id"]
				attr.Optional = true
				attr.Computed = true
				attr.MarkdownDescription += idNameRemark
				return attr
			}(),
			"name": func() tfsdk.Attribute {
				attr := attributes["name"]
				attr.Optional = true
				attr.Computed = true
				return attr
			}(),
			"database_id": func() tfsdk.Attribute {
				attr := attributes["database_id"]
				attr.Optional = true
				attr.Computed = true
				return attr
			}(),
			"owner_id": func() tfsdk.Attribute {
				attr := attributes["owner_id"]
				attr.Computed = true
				return attr
			}(),
		},
	}
}

func (d dataSource) Read(ctx context.Context, req datasource.ReadRequest[resourceData], resp *datasource.ReadResponse[resourceData]) {
	var (
		db     sql.Database
		schema sql.Schema
	)

	schemaId := common.ParseDbObjectId[sql.SchemaId](ctx, req.Config.Id.Value)

	utils.StopOnError(ctx).
		Then(func() {
			if schemaId.IsEmpty {
				db = common.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.Value)
			} else {
				db = sql.GetDatabase(ctx, req.Conn, schemaId.DbId)
			}
		}).
		Then(func() {
			if schemaId.IsEmpty {
				schema = sql.GetSchemaByName(ctx, db, req.Config.Name.Value)
			} else {
				schema = sql.GetSchema(ctx, db, schemaId.ObjectId)
			}
		}).
		Then(func() {
			resp.SetState(req.Config.withSchemaData(ctx, schema))
		})
}
