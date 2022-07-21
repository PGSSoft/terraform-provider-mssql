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
	_ tfsdk.ResourceType            = AzureADUserResourceType{}
	_ tfsdk.Resource                = azureADUserResource{}
	_ tfsdk.ResourceWithImportState = azureADUserResource{}
)

type AzureADUserResourceType struct{}

func (r AzureADUserResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: `
Managed database-level user mapped to Azure AD identity (user, group, service principal or managed identity).

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
					tfsdk.RequiresReplace(),
				}

				return attr
			}(),
		},
	}, nil
}

func (r AzureADUserResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) azureADUserResource {
		return azureADUserResource{Resource: base}
	})
}

type azureADUserResource struct {
	Resource
}

func (r azureADUserResource) Create(ctx context.Context, request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
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

func (r azureADUserResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {
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

func (r azureADUserResource) Update(context.Context, tfsdk.UpdateResourceRequest, *tfsdk.UpdateResourceResponse) {
	panic("not supported")
}

func (r azureADUserResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
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

func (r azureADUserResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), request, response)
}
