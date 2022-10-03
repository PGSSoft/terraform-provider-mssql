package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sqlLoginData struct {
	BaseDataSource
}

func (p mssqlProvider) NewSqlLoginDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[sqlLoginDataSourceData](&sqlLoginData{})
	}
}

func (d *sqlLoginData) GetName() string {
	return "sql_login"
}

func (d *sqlLoginData) GetSchema(context.Context) tfsdk.Schema {
	a := map[string]tfsdk.Attribute{}
	for n, attribute := range sqlLoginAttributes {
		attribute.Required = n == "name"
		attribute.Computed = n != "name"
		a[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about single SQL login.",
		Attributes:  a,
	}
}

func (d *sqlLoginData) Read(ctx context.Context, req datasource.ReadRequest[sqlLoginDataSourceData], resp *datasource.ReadResponse[sqlLoginDataSourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() {
			login = sql.GetSqlLoginByName(ctx, d.conn, req.Config.Name.Value)

			if login == nil || !login.Exists(ctx) {
				utils.AddError(ctx, "Login does not exist", fmt.Errorf("could not find SQL Login '%s'", req.Config.Name.Value))
			}
		}).
		Then(func() {
			state := req.Config.withSettings(login.GetSettings(ctx))
			state.Id = types.String{Value: fmt.Sprint(login.GetId(ctx))}

			resp.SetState(state)
		})
}
