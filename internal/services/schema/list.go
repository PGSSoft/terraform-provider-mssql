package schema

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceData struct {
	Id         types.String   `tfsdk:"id"`
	DatabaseId types.String   `tfsdk:"database_id"`
	Schemas    []resourceData `tfsdk:"schemas"`
}

type listDataSource struct{}

func (l listDataSource) GetName() string {
	return "schemas"
}

func (l listDataSource) GetSchema(context.Context) tfsdk.Schema {
	schemaAttrs := map[string]tfsdk.Attribute{}

	for n, attr := range attributes {
		a := attr
		a.Computed = true
		schemaAttrs[n] = a
	}

	return tfsdk.Schema{
		MarkdownDescription: "Obtains information about all schemas found in SQL database.",
		Attributes: map[string]tfsdk.Attribute{
			"id": common.ToResourceId(tfsdk.Attribute{
				MarkdownDescription: "ID of the data source, equals to database ID",
				Type:                types.StringType,
			}),
			"database_id": common.DatabaseIdResourceAttribute,
			"schemas": {
				MarkdownDescription: "Set of schemas found in the DB.",
				Attributes:          tfsdk.SetNestedAttributes(schemaAttrs),
				Computed:            true,
			},
		},
	}
}

func (l listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	var schemas map[sql.SchemaId]sql.Schema
	var dbId sql.DatabaseId

	db := common.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString())

	req.
		Then(func() {
			dbId = db.GetId(ctx)
			schemas = sql.GetSchemas(ctx, db)
		}).
		Then(func() {
			data := listDataSourceData{
				DatabaseId: types.StringValue(fmt.Sprint(dbId)),
				Schemas:    []resourceData{},
			}
			data.Id = data.DatabaseId

			for _, schema := range schemas {
				data.Schemas = append(data.Schemas, resourceData{}.withSchemaData(ctx, schema))
			}

			resp.SetState(data)
		})
}
