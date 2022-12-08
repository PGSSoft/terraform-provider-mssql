package script

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceData struct {
	Id         types.String        `tfsdk:"id"`
	DatabaseId types.String        `tfsdk:"database_id"`
	Query      types.String        `tfsdk:"query"`
	Result     []map[string]string `tfsdk:"result"`
}

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "query"
}

func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Retrieves arbitrary SQL query result.

-> **Note** This data source is meant to be an escape hatch for all cases not supported by the provider's data sources. Whenever possible, use dedicated data sources, which offer better plan, validation and error reporting.
`

	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Used only internally by Terraform. Always set to `query`",
			Computed:            true,
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common.AttributeDescriptions["database_id"],
			Required:            true,
		},
		"query": schema.StringAttribute{
			MarkdownDescription: "SQL query returning single result set, with any number of rows, where all columns are strings",
			Required:            true,
		},
		"result": schema.ListAttribute{
			MarkdownDescription: "Results of the SQL query, represented as list of maps, where the map key corresponds to column name and the value is the value of column in given row.",
			Computed:            true,
			ElementType:         types.MapType{ElemType: types.StringType},
		},
	}
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[dataSourceData], resp *datasource.ReadResponse[dataSourceData]) {
	var (
		db     sql.Database
		result []map[string]string
	)

	req.
		Then(func() { db = common.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString()) }).
		Then(func() { result = db.Query(ctx, req.Config.Query.ValueString()) }).
		Then(func() {
			req.Config.Result = result
			req.Config.Id = types.StringValue("query")
			resp.SetState(req.Config)
		})
}
