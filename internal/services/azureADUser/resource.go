package azureADUser

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"

	"github.com/PGSSoft/terraform-provider-mssql/internal/planModifiers"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type res struct{}

func (r *res) GetName() string {
	return "azuread_user"
}

func (r *res) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: `
Managed database-level user mapped to Azure AD identity (user or group).

-> **Note** When using this resource, Azure SQL server managed identity does not need any [AzureAD role assignments](https://docs.microsoft.com/en-us/azure/azure-sql/database/authentication-aad-service-principal?view=azuresql).
`,
		Attributes: map[string]tfsdk.Attribute{
			"id":          common2.ToResourceId(attributes["id"]),
			"name":        common2.ToRequiredImmutable(attributes["name"]),
			"database_id": common2.ToRequiredImmutable(attributes["database_id"]),
			"user_object_id": func() tfsdk.Attribute {
				attr := attributes["user_object_id"]
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

func (r *res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.Value) }).
		Then(func() { user = sql.CreateUser(ctx, db, req.Plan.toSettings()) }).
		Then(func() {
			req.Plan.Id = types.String{Value: common2.DbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String()}
		}).
		Then(func() { resp.State = req.Plan })
}

func (r *res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var (
		id   common2.DbObjectId[sql.UserId]
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { id = common2.ParseDbObjectId[sql.UserId](ctx, req.State.Id.Value) }).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, id.DbId) }).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() {
			state := req.State.withSettings(ctx, user.GetSettings(ctx))
			state.DatabaseId = types.String{Value: fmt.Sprint(id.DbId)}
			resp.SetState(state)
		})
}

func (r *res) Update(context.Context, resource.UpdateRequest[resourceData], *resource.UpdateResponse[resourceData]) {
	panic("not supported")
}

func (r *res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	var (
		id   common2.DbObjectId[sql.UserId]
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() {
			db = common2.GetResourceDb(ctx, req.Conn, req.State.DatabaseId.Value)
			id = common2.ParseDbObjectId[sql.UserId](ctx, req.State.Id.Value)
		}).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() { user.Drop(ctx) })
}
