package schema

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (l listDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about all schemas found in SQL database."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "ID of the data source, equals to database ID",
			Computed:            true,
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common.AttributeDescriptions["database_id"] + " Defaults to ID of `master`.",
			Optional:            true,
			Computed:            true,
		},
		"schemas": schema.SetNestedAttribute{
			MarkdownDescription: "Set of schemas found in the DB.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: attrDescriptions["id"],
						Computed:            true,
					},
					"database_id": schema.StringAttribute{
						MarkdownDescription: common.AttributeDescriptions["database_id"],
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: attrDescriptions["name"],
						Computed:            true,
					},
					"owner_id": schema.StringAttribute{
						MarkdownDescription: attrDescriptions["owner_id"],
						Computed:            true,
					},
				},
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
