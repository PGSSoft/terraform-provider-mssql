package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/resource"

	"github.com/PGSSoft/terraform-provider-mssql/internal/planModifiers"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type azureADServicePrincipalResource struct {
	BaseResource
}

func (p mssqlProvider) NewAzureADServicePrincipalResource() func() sdkresource.Resource {
	return func() sdkresource.Resource {
		return resource.WrapResource[azureADServicePrincipalResourceData](&azureADServicePrincipalResource{})
	}
}

func (r *azureADServicePrincipalResource) GetName() string {
	return "azuread_service_principal"
}

func (r *azureADServicePrincipalResource) GetSchema(context.Context) tfsdk.Schema {
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
					sdkresource.RequiresReplace(),
				}

				return attr
			}(),
		},
	}
}

func (r *azureADServicePrincipalResource) Create(ctx context.Context, req resource.CreateRequest[azureADServicePrincipalResourceData], resp *resource.CreateResponse[azureADServicePrincipalResourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { db = getResourceDb(ctx, r.conn, req.Plan.DatabaseId.Value) }).
		Then(func() { user = sql.CreateUser(ctx, db, req.Plan.toSettings()) }).
		Then(func() {
			req.Plan.Id = types.String{Value: dbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
		}).
		Then(func() { resp.State = req.Plan })
}

func (r *azureADServicePrincipalResource) Read(ctx context.Context, req resource.ReadRequest[azureADServicePrincipalResourceData], resp *resource.ReadResponse[azureADServicePrincipalResourceData]) {
	var (
		id   dbObjectId[sql.UserId]
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { id = parseDbObjectId[sql.UserId](ctx, req.State.Id.Value) }).
		Then(func() { db = sql.GetDatabase(ctx, r.conn, id.DbId) }).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() {
			state := req.State.withSettings(ctx, user.GetSettings(ctx))
			state.DatabaseId = types.String{Value: fmt.Sprint(id.DbId)}
			resp.SetState(state)
		})
}

func (r *azureADServicePrincipalResource) Update(context.Context, resource.UpdateRequest[azureADServicePrincipalResourceData], *resource.UpdateResponse[azureADServicePrincipalResourceData]) {
	panic("not supported")
}

func (r *azureADServicePrincipalResource) Delete(ctx context.Context, req resource.DeleteRequest[azureADServicePrincipalResourceData], _ *resource.DeleteResponse[azureADServicePrincipalResourceData]) {
	var (
		id   dbObjectId[sql.UserId]
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() {
			db = getResourceDb(ctx, r.conn, req.State.DatabaseId.Value)
			id = parseDbObjectId[sql.UserId](ctx, req.State.Id.Value)
		}).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() { user.Drop(ctx) })
}
