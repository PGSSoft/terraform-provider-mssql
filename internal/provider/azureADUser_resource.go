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

type azureADUserResource struct {
	BaseResource
}

func (p mssqlProvider) NewAzureADUserResource() func() sdkresource.Resource {
	return func() sdkresource.Resource {
		return resource.WrapResource[azureADUserResourceData](&azureADUserResource{})
	}
}

func (r *azureADUserResource) GetName() string {
	return "azuread_user"
}

func (r *azureADUserResource) GetSchema(context.Context) tfsdk.Schema {
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
					sdkresource.RequiresReplace(),
				}

				return attr
			}(),
		},
	}
}

func (r *azureADUserResource) Create(ctx context.Context, req resource.CreateRequest[azureADUserResourceData], resp *resource.CreateResponse[azureADUserResourceData]) {
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

func (r *azureADUserResource) Read(ctx context.Context, req resource.ReadRequest[azureADUserResourceData], resp *resource.ReadResponse[azureADUserResourceData]) {
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

func (r *azureADUserResource) Update(context.Context, resource.UpdateRequest[azureADUserResourceData], *resource.UpdateResponse[azureADUserResourceData]) {
	panic("not supported")
}

func (r *azureADUserResource) Delete(ctx context.Context, req resource.DeleteRequest[azureADUserResourceData], resp *resource.DeleteResponse[azureADUserResourceData]) {
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
