package provider

import (
	"context"
	"fmt"

	"github.com/PGSSoft/terraform-provider-mssql/internal/planModifiers"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.ResourceType            = AzureADServicePrincipalResourceType{}
	_ tfsdk.Resource                = azureADServicePrincipalResource{}
	_ tfsdk.ResourceWithImportState = azureADServicePrincipalResource{}
)

type AzureADServicePrincipalResourceType struct{}

func (r AzureADServicePrincipalResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
					tfsdk.RequiresReplace(),
				}

				return attr
			}(),
		},
	}, nil
}

func (r AzureADServicePrincipalResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) azureADServicePrincipalResource {
		return azureADServicePrincipalResource{Resource: base}
	})
}

type azureADServicePrincipalResource struct {
	Resource
}

func (r azureADServicePrincipalResource) Create(ctx context.Context, request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
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

func (r azureADServicePrincipalResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {
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

func (r azureADServicePrincipalResource) Update(context.Context, tfsdk.UpdateResourceRequest, *tfsdk.UpdateResourceResponse) {
	panic("not supported")
}

func (r azureADServicePrincipalResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
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

func (r azureADServicePrincipalResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), request, response)
}
