package provider

import (
	"context"
	"fmt"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ datasource.DataSourceWithConfigure = &databaseData{}
)

type databaseData struct {
	Resource
}

func (p mssqlProvider) NewDatabaseDataSource() func() datasource.DataSource {
	return func() datasource.DataSource {
		return &databaseData{}
	}
}

func (s *databaseData) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	s.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (s databaseData) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "mssql_database"
}

func (d databaseData) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (d databaseData) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data, _ := getDB(ctx, request.Config)
	if utils.HasError(ctx) {
		return
	}

	db := sql.GetDatabaseByName(ctx, d.Db, data.Name.Value)

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
