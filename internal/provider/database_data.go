package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tkielar/terraform-provider-mssql/internal/utils"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.DataSourceType = DatabaseDataSourceType{}
	_ tfsdk.DataSource     = databaseData{}
)

type DatabaseDataSourceType struct{}

func (d DatabaseDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	a := map[string]tfsdk.Attribute{}
	for n, attribute := range databaseAttributes {
		attribute.Required = n == "name"
		attribute.Computed = n != "name"
		a[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about single database.",
		Attributes:  a,
	}, nil
}

func (d DatabaseDataSourceType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) databaseData {
		return databaseData{Resource: base}
	})
}

type databaseData struct {
	Resource
}

func (d databaseData) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data, _ := getDB(ctx, request.Config)
	if utils.HasError(ctx) {
		return
	}

	db := d.Db.GetDatabaseByName(ctx, data.Name.Value)

	if db == nil || !db.Exists(ctx) {
		response.State.RemoveResource(ctx)
		utils.AddError(ctx, "DB does not exist", fmt.Errorf("could not find DB '%s'", data.Name.Value))
	}

	if utils.HasError(ctx) {
		return
	}

	dbSettings := db.GetSettings(ctx)
	if utils.HasError(ctx) {
		return
	}

	data = data.withSettings(dbSettings)

	if data.Id.Unknown || data.Id.Null {
		data.Id = types.String{Value: fmt.Sprint(db.GetId(ctx))}
	}

	utils.SetData(ctx, &response.State, data)
}
