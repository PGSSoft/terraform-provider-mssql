package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.DataSourceType = AzureADServicePrincipalDataSourceType{}
	_ tfsdk.DataSource     = azureADServicePrincipalData{}
)

type AzureADServicePrincipalDataSourceType struct{}

func (s AzureADServicePrincipalDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
	}, nil
}

func (s AzureADServicePrincipalDataSourceType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) azureADServicePrincipalData {
		return azureADServicePrincipalData{Resource: base}
	})
}

type azureADServicePrincipalData struct {
	Resource
}

func (s azureADServicePrincipalData) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	var (
		data azureADServicePrincipalResourceData
		db   sql.Database
		user sql.User
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[azureADServicePrincipalResourceData](ctx, request.Config) }).
		Then(func() { db = getResourceDb(ctx, s.Db, data.DatabaseId.Value) }).
		Then(func() {
			if !data.Name.IsNull() && !data.Name.IsUnknown() {
				user = sql.GetUserByName(ctx, db, data.Name.Value)
				return
			}

			for _, u := range sql.GetUsers(ctx, db) {
				settings := u.GetSettings(ctx)
				if settings.Type == sql.USER_TYPE_AZUREAD && strings.EqualFold(fmt.Sprint(settings.AADObjectId), data.ClientId.Value) {
					user = u
					return
				}
			}

			utils.AddError(ctx, "User does not exist", fmt.Errorf("could not find user with name=%q and client_id=%q", data.Name.Value, data.ClientId.Value))
		}).
		Then(func() {
			data.Id = types.String{Value: dbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
			data = data.withSettings(ctx, user.GetSettings(ctx))
		}).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}
