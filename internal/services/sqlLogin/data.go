package sqlLogin

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "sql_login"
}

func (d *dataSource) GetSchema(context.Context) tfsdk.Schema {
	a := map[string]tfsdk.Attribute{}
	for n, attribute := range attributes {
		attribute.Required = n == "name"
		attribute.Computed = n != "name"
		a[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about single SQL login.",
		Attributes:  a,
	}
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[dataSourceData], resp *datasource.ReadResponse[dataSourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() {
			login = sql.GetSqlLoginByName(ctx, req.Conn, req.Config.Name.ValueString())

			if login == nil || !login.Exists(ctx) {
				utils.AddError(ctx, "Login does not exist", fmt.Errorf("could not find SQL Login '%s'", req.Config.Name.ValueString()))
			}
		}).
		Then(func() {
			state := req.Config.withSettings(login.GetSettings(ctx))
			state.Id = types.StringValue(fmt.Sprint(login.GetId(ctx)))

			resp.SetState(state)
		})
}
