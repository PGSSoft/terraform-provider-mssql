package azureADUser

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "azuread_user"
}

func (d *dataSource) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range attributes {
		attr.Required = n == "database_id"
		attr.Optional = n != "id" && !attr.Required
		attr.Computed = !attr.Required
		attrs[n] = attr
	}

	return tfsdk.Schema{
		Description: "Obtains information about single Azure AD database user.",
		Attributes:  attrs,
	}
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[resourceData], resp *datasource.ReadResponse[resourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.Value) }).
		Then(func() {
			if !req.Config.Name.IsNull() && !req.Config.Name.IsUnknown() {
				user = sql.GetUserByName(ctx, db, req.Config.Name.Value)
				return
			}

			for _, u := range sql.GetUsers(ctx, db) {
				settings := u.GetSettings(ctx)
				if settings.Type == sql.USER_TYPE_AZUREAD && strings.ToUpper(fmt.Sprint(settings.AADObjectId)) == strings.ToUpper(req.Config.UserObjectId.Value) {
					user = u
					return
				}
			}

			utils.AddError(ctx, "User does not exist", fmt.Errorf("could not find user with name=%q and object_id=%q", req.Config.Name.Value, req.Config.UserObjectId.Value))
		}).
		Then(func() {
			state := req.Config.withSettings(ctx, user.GetSettings(ctx))
			state.Id = types.String{Value: common2.DbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
			resp.SetState(state)
		})
}
