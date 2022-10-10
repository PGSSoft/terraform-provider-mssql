package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type azureADServicePrincipalData struct {
	BaseDataSource
}

func (p mssqlProvider) NewAzureADServicePrincipalDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[azureADServicePrincipalResourceData](&azureADServicePrincipalData{})
	}
}

func (d *azureADServicePrincipalData) GetName() string {
	return "azuread_service_principal"
}

func (d *azureADServicePrincipalData) GetSchema(_ context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range azureADServicePrincipalAttributes {
		attr.Required = n == "database_id"
		attr.Optional = n != "id" && !attr.Required
		attr.Computed = !attr.Required
		attrs[n] = attr
	}

	return tfsdk.Schema{
		Description: "Obtains information about single Azure AD Service Principal database user.",
		Attributes:  attrs,
	}
}

func (d *azureADServicePrincipalData) Read(ctx context.Context, request datasource.ReadRequest[azureADServicePrincipalResourceData], response *datasource.ReadResponse[azureADServicePrincipalResourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	request.
		Then(func() { db = getResourceDb(ctx, d.conn, request.Config.DatabaseId.Value) }).
		Then(func() {
			if !request.Config.Name.IsNull() && !request.Config.Name.IsUnknown() {
				user = sql.GetUserByName(ctx, db, request.Config.Name.Value)
				return
			}

			for _, u := range sql.GetUsers(ctx, db) {
				settings := u.GetSettings(ctx)
				if settings.Type == sql.USER_TYPE_AZUREAD && strings.EqualFold(fmt.Sprint(settings.AADObjectId), request.Config.ClientId.Value) {
					user = u
					return
				}
			}

			utils.AddError(ctx, "User does not exist", fmt.Errorf("could not find user with name=%q and client_id=%q", request.Config.Name.Value, request.Config.ClientId.Value))
		}).
		Then(func() {
			state := request.Config
			state.Id = types.String{Value: dbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
			state = state.withSettings(ctx, user.GetSettings(ctx))
			response.SetState(state)
		})
}
