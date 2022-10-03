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

type databaseData struct {
	BaseDataSource
}

func (p mssqlProvider) NewDatabaseDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[databaseResourceData](&databaseData{})
	}
}

func (d *databaseData) GetName() string {
	return "database"
}

func (d *databaseData) GetSchema(context.Context) tfsdk.Schema {
	a := map[string]tfsdk.Attribute{}
	for n, attribute := range databaseAttributes {
		attribute.Required = n == "name"
		attribute.Computed = n != "name"
		a[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about single database.",
		Attributes:  a,
	}
}

func (d *databaseData) Read(ctx context.Context, req datasource.ReadRequest[databaseResourceData], resp *datasource.ReadResponse[databaseResourceData]) {
	var db sql.Database

	req.
		Then(func() {
			db = sql.GetDatabaseByName(ctx, d.conn, req.Config.Name.Value)

			if db == nil || !db.Exists(ctx) {
				utils.AddError(ctx, "DB does not exist", fmt.Errorf("could not find DB '%s'", req.Config.Name.Value))
			}
		}).
		Then(func() {
			state := req.Config.withSettings(db.GetSettings(ctx))

			if state.Id.Unknown || state.Id.Null {
				state.Id = types.String{Value: fmt.Sprint(db.GetId(ctx))}
			}

			resp.SetState(state)
		})
}
