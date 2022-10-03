package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type azureADUserData struct {
	BaseDataSource
}

func (p mssqlProvider) NewAzureADUserDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[azureADUserResourceData](&azureADUserData{})
	}
}

func (d *azureADUserData) GetName() string {
	return "azuread_user"
}

func (d *azureADUserData) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range azureADUserAttributes {
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

func (d *azureADUserData) Read(ctx context.Context, req datasource.ReadRequest[azureADUserResourceData], resp *datasource.ReadResponse[azureADUserResourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { db = getResourceDb(ctx, d.conn, req.Config.DatabaseId.Value) }).
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
			state.Id = types.String{Value: dbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
			resp.SetState(state)
		})
}
