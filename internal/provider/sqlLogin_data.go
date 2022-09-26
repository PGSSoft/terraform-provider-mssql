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
	_ datasource.DataSourceWithConfigure = &sqlLoginData{}
)

type sqlLoginData struct {
	Resource
}

func (p mssqlProvider) NewSqlLoginDataSource() func() datasource.DataSource {
	return func() datasource.DataSource {
		return &sqlLoginData{}
	}
}

func (s *sqlLoginData) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	s.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (s sqlLoginData) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "mssql_sql_login"
}

func (d sqlLoginData) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	a := map[string]tfsdk.Attribute{}
	for n, attribute := range sqlLoginAttributes {
		attribute.Required = n == "name"
		attribute.Computed = n != "name"
		a[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about single SQL login.",
		Attributes:  a,
	}, nil
}

func (d sqlLoginData) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlLoginDataSourceData](ctx, request.Config)

	login := sql.GetSqlLoginByName(ctx, d.Db, data.Name.Value)

	if login == nil || !login.Exists(ctx) {
		response.State.RemoveResource(ctx)
		utils.AddError(ctx, "Login does not exist", fmt.Errorf("could not find SQL Login '%s'", data.Name.Value))
	}

	if utils.HasError(ctx) {
		return
	}

	if data = data.withSettings(login.GetSettings(ctx)); utils.HasError(ctx) {
		return
	}

	data.Id = types.String{Value: fmt.Sprint(login.GetId(ctx))}

	utils.SetData(ctx, &response.State, data)
}
