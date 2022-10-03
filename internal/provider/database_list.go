package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type databaseListData struct {
	Id        types.String           `tfsdk:"id"`
	Databases []databaseResourceData `tfsdk:"databases"`
}

type databaseList struct {
	BaseDataSource
}

func (p mssqlProvider) NewDatabaseListDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[databaseListData](&databaseList{})
	}
}

func (l *databaseList) GetName() string {
	return "databases"
}

func (l *databaseList) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attribute := range databaseAttributes {
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

func (l *databaseList) Read(ctx context.Context, req datasource.ReadRequest[databaseListData], resp *datasource.ReadResponse[databaseListData]) {
	var dbs map[sql.DatabaseId]sql.Database

	req.
		Then(func() { dbs = sql.GetDatabases(ctx, l.conn) }).
		Then(func() {
			result := databaseListData{
				Id:        types.String{Value: ""},
				Databases: []databaseResourceData{},
			}

			for id, db := range dbs {
				r := databaseResourceData{
					Id: types.String{Value: fmt.Sprint(id)},
				}
				result.Databases = append(result.Databases, r.withSettings(db.GetSettings(ctx)))
			}

			resp.SetState(result)
		})
}
