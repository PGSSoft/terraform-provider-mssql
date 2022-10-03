package provider

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type sqlUserData struct {
	BaseDataSource
}

func (p mssqlProvider) NewSqlUserDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[sqlUserResourceData](&sqlUserData{})
	}
}

func (d *sqlUserData) GetName() string {
	return "sql_user"
}

func (d *sqlUserData) GetSchema(context.Context) tfsdk.Schema {
	a := map[string]tfsdk.Attribute{}
	for n, attribute := range sqlUserAttributes {
		attribute.Required = n == "name"
		attribute.Optional = n == "database_id"
		attribute.Computed = n != "name"
		a[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about single SQL database user.",
		Attributes:  a,
	}
}

func (d *sqlUserData) Read(ctx context.Context, req datasource.ReadRequest[sqlUserResourceData], resp *datasource.ReadResponse[sqlUserResourceData]) {
	var db sql.Database
	var user sql.User

	req.
		Then(func() { db = getResourceDb(ctx, d.conn, req.Config.DatabaseId.Value) }).
		Then(func() { user = sql.GetUserByName(ctx, db, req.Config.Name.Value) }).
		Then(func() {
			state := req.Config.withIds(db.GetId(ctx), user.GetId(ctx))
			resp.SetState(state.withSettings(user.GetSettings(ctx)))
		})
}
