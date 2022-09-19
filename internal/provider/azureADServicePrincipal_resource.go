package provider

import (
	"context"
	"fmt"

	"github.com/PGSSoft/terraform-provider-mssql/internal/planModifiers"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ resource.ResourceWithImportState = azureADServicePrincipalResource{}
	_ resource.ResourceWithConfigure   = &azureADServicePrincipalResource{}
)

type azureADServicePrincipalResource struct {
	Resource
}

func (p mssqlProvider) NewAzureADServicePrincipalResource() func() resource.Resource {
	return func() resource.Resource {
		return &azureADServicePrincipalResource{}
	}
}

func (s azureADServicePrincipalResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "mssql_azuread_service_principal"
}

func (r *azureADServicePrincipalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (r azureADServicePrincipalResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: `
Managed database-level user mapped to Azure AD identity (service principal or managed identity).

-> **Note** When using this resource, Azure SQL server managed identity does not need any [AzureAD role assignments](https://docs.microsoft.com/en-us/azure/azure-sql/database/authentication-aad-service-principal?view=azuresql).
`,
		Attributes: map[string]tfsdk.Attribute{
			"id":          toResourceId(azureADServicePrincipalAttributes["id"]),
			"name":        toRequiredImmutable(azureADServicePrincipalAttributes["name"]),
			"database_id": toRequiredImmutable(azureADServicePrincipalAttributes["database_id"]),
			"client_id": func() tfsdk.Attribute {
				attr := azureADServicePrincipalAttributes["client_id"]
				attr.Required = true
				attr.PlanModifiers = tfsdk.AttributePlanModifiers{
					planModifiers.IgnoreCase(),
					resource.RequiresReplace(),
				}

				return attr
			}(),
		},
	}, nil
}

func (r azureADServicePrincipalResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		data azureADServicePrincipalResourceData
		db   sql.Database
		user sql.User
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[azureADServicePrincipalResourceData](ctx, request.Plan) }).
		Then(func() { db = getResourceDb(ctx, r.Db, data.DatabaseId.Value) }).
		Then(func() { user = sql.CreateUser(ctx, db, data.toSettings()) }).
		Then(func() {
			data.Id = types.String{Value: dbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
		}).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}

func (r azureADServicePrincipalResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		data azureADServicePrincipalResourceData
		id   dbObjectId[sql.UserId]
		db   sql.Database
		user sql.User
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[azureADServicePrincipalResourceData](ctx, request.State) }).
		Then(func() { id = parseDbObjectId[sql.UserId](ctx, data.Id.Value) }).
		Then(func() { db = sql.GetDatabase(ctx, r.Db, id.DbId) }).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() {
			data = data.withSettings(ctx, user.GetSettings(ctx))
			data.DatabaseId = types.String{Value: fmt.Sprint(id.DbId)}
		}).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}

func (r azureADServicePrincipalResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("not supported")
}

func (r azureADServicePrincipalResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		data azureADServicePrincipalResourceData
		id   dbObjectId[sql.UserId]
		db   sql.Database
		user sql.User
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[azureADServicePrincipalResourceData](ctx, request.State) }).
		Then(func() {
			db = getResourceDb(ctx, r.Db, data.DatabaseId.Value)
			id = parseDbObjectId[sql.UserId](ctx, data.Id.Value)
		}).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() { user.Drop(ctx) }).
		Then(func() { response.State.RemoveResource(ctx) })
}

func (r azureADServicePrincipalResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
