package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ datasource.DataSourceWithConfigure = &azureADServicePrincipalData{}
)

type azureADServicePrincipalData struct {
	Resource
}

func (p mssqlProvider) NewAzureADServicePrincipalDataSource() func() datasource.DataSource {
	return func() datasource.DataSource {
		return azureADServicePrincipalData{}
	}
}

func (s *azureADServicePrincipalData) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	s.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (s azureADServicePrincipalData) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "mssql_azuread_service_principal"
}

func (s azureADServicePrincipalData) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (s azureADServicePrincipalData) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
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
