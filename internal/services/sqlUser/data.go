package sqlUser

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "sql_user"
}

func (d *dataSource) GetSchema(context.Context) tfsdk.Schema {
	a := map[string]tfsdk.Attribute{}
	for n, attribute := range attributes {
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

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[resourceData], resp *datasource.ReadResponse[resourceData]) {
	var db sql.Database
	var user sql.User

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString()) }).
		Then(func() { user = sql.GetUserByName(ctx, db, req.Config.Name.ValueString()) }).
		Then(func() {
			state := req.Config.withIds(db.GetId(ctx), user.GetId(ctx))
			resp.SetState(state.withSettings(user.GetSettings(ctx)))
		})
}
