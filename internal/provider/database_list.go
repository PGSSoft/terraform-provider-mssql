package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.DataSourceType = DatabaseListDataSourceType{}
	_ tfsdk.DataSource     = databaseList{}
)

type DatabaseListDataSourceType struct{}

func (l DatabaseListDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
	}, nil
}

func (l DatabaseListDataSourceType) NewDataSource(ctx context.Context, provider tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return newResource(ctx, provider, func(base Resource) databaseList {
		return databaseList{Resource: base}
	})
}

type databaseList struct {
	Resource
}

func (l databaseList) Read(ctx context.Context, _ tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	dbs := sql.GetDatabases(ctx, l.Db)
	if utils.HasError(ctx) {
		return
	}

	result := databaseListData{
		Id:        types.String{Value: ""},
		Databases: []databaseResourceData{},
	}

	for id, db := range dbs {
		s := db.GetSettings(ctx)

		if utils.HasError(ctx) {
			return
		}

		r := databaseResourceData{Id: types.String{Value: fmt.Sprint(id)}}
		result.Databases = append(result.Databases, r.withSettings(s))
	}

	utils.SetData(ctx, &response.State, result)
}

type databaseListData struct {
	Id        types.String           `tfsdk:"id"`
	Databases []databaseResourceData `tfsdk:"databases"`
}
