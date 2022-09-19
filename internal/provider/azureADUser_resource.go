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
	_ resource.ResourceWithConfigure   = &azureADUserResource{}
	_ resource.ResourceWithImportState = azureADUserResource{}
)

type azureADUserResource struct {
	Resource
}

func (p mssqlProvider) NewAzureADUserResource() func() resource.Resource {
	return func() resource.Resource {
		return &azureADUserResource{}
	}
}

func (s azureADUserResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "mssql_azuread_user"
}

func (r *azureADUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (r azureADUserResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: `
Managed database-level user mapped to Azure AD identity (user or group).

-> **Note** When using this resource, Azure SQL server managed identity does not need any [AzureAD role assignments](https://docs.microsoft.com/en-us/azure/azure-sql/database/authentication-aad-service-principal?view=azuresql).
`,
		Attributes: map[string]tfsdk.Attribute{
			"id":          toResourceId(azureADUserAttributes["id"]),
			"name":        toRequiredImmutable(azureADUserAttributes["name"]),
			"database_id": toRequiredImmutable(azureADUserAttributes["database_id"]),
			"user_object_id": func() tfsdk.Attribute {
				attr := azureADUserAttributes["user_object_id"]
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

func (r azureADUserResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		data azureADUserResourceData
		db   sql.Database
		user sql.User
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[azureADUserResourceData](ctx, request.Plan) }).
		Then(func() { db = getResourceDb(ctx, r.Db, data.DatabaseId.Value) }).
		Then(func() { user = sql.CreateUser(ctx, db, data.toSettings()) }).
		Then(func() {
			data.Id = types.String{Value: dbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
		}).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}

func (r azureADUserResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		data azureADUserResourceData
		id   dbObjectId[sql.UserId]
		db   sql.Database
		user sql.User
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[azureADUserResourceData](ctx, request.State) }).
		Then(func() { id = parseDbObjectId[sql.UserId](ctx, data.Id.Value) }).
		Then(func() { db = sql.GetDatabase(ctx, r.Db, id.DbId) }).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() {
			data = data.withSettings(ctx, user.GetSettings(ctx))
			data.DatabaseId = types.String{Value: fmt.Sprint(id.DbId)}
		}).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}

func (r azureADUserResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("not supported")
}

func (r azureADUserResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		data azureADUserResourceData
		id   dbObjectId[sql.UserId]
		db   sql.Database
		user sql.User
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[azureADUserResourceData](ctx, request.State) }).
		Then(func() {
			db = getResourceDb(ctx, r.Db, data.DatabaseId.Value)
			id = parseDbObjectId[sql.UserId](ctx, data.Id.Value)
		}).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() { user.Drop(ctx) }).
		Then(func() { response.State.RemoveResource(ctx) })
}

func (r azureADUserResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
