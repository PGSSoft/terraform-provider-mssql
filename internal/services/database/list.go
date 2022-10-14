package database

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceData struct {
	Id        types.String   `tfsdk:"id"`
	Databases []resourceData `tfsdk:"databases"`
}

type listDataSource struct{}

func (l *listDataSource) GetName() string {
	return "databases"
}

func (l *listDataSource) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attribute := range attributes {
		attribute.Computed = true
		attrs[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about all databases found in SQL Server instance.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "ID of the resource used only internally by the provider.",
			},
			"databases": {
				Description: "Set of database objects",
				Attributes:  tfsdk.SetNestedAttributes(attrs),
				Computed:    true,
			},
		},
	}
}

func (l *listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	var dbs map[sql.DatabaseId]sql.Database

	req.
		Then(func() { dbs = sql.GetDatabases(ctx, req.Conn) }).
		Then(func() {
			result := listDataSourceData{
				Id:        types.String{Value: ""},
				Databases: []resourceData{},
			}

			for id, db := range dbs {
				r := resourceData{
					Id: types.String{Value: fmt.Sprint(id)},
				}
				result.Databases = append(result.Databases, r.withSettings(db.GetSettings(ctx)))
			}

			resp.SetState(result)
		})
}
