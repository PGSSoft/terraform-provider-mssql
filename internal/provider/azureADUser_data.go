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
	_ datasource.DataSourceWithConfigure = &azureADUserData{}
)

type azureADUserData struct {
	Resource
}

func (p mssqlProvider) NewAzureADUserDataSource() func() datasource.DataSource {
	return func() datasource.DataSource {
		return azureADUserData{}
	}
}

func (s *azureADUserData) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	s.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (s azureADUserData) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "mssql_azuread_user"
}

func (s azureADUserData) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
	}, nil
}

func (s azureADUserData) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		data azureADUserResourceData
		db   sql.Database
		user sql.User
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[azureADUserResourceData](ctx, request.Config) }).
		Then(func() { db = getResourceDb(ctx, s.Db, data.DatabaseId.Value) }).
		Then(func() {
			if !data.Name.IsNull() && !data.Name.IsUnknown() {
				user = sql.GetUserByName(ctx, db, data.Name.Value)
				return
			}

			for _, u := range sql.GetUsers(ctx, db) {
				settings := u.GetSettings(ctx)
				if settings.Type == sql.USER_TYPE_AZUREAD && strings.ToUpper(fmt.Sprint(settings.AADObjectId)) == strings.ToUpper(data.UserObjectId.Value) {
					user = u
					return
				}
			}

			utils.AddError(ctx, "User does not exist", fmt.Errorf("could not find user with name=%q and object_id=%q", data.Name.Value, data.UserObjectId.Value))
		}).
		Then(func() {
			data.Id = types.String{Value: dbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
			data = data.withSettings(ctx, user.GetSettings(ctx))
		}).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}
