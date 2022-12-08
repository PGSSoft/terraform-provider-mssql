package azureADServicePrincipal

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/PGSSoft/terraform-provider-mssql/internal/planModifiers"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type res struct{}

func (r *res) GetName() string {
	return "azuread_service_principal"
}

func (r *res) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = `
Managed database-level user mapped to Azure AD identity (service principal or managed identity).

-> **Note** When using this resource, Azure SQL server managed identity does not need any [AzureAD role assignments](https://docs.microsoft.com/en-us/azure/azure-sql/database/authentication-aad-service-principal?view=azuresql).
`
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["id"],
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["name"],
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: validators.UserNameValidators,
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common2.AttributeDescriptions["database_id"],
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"client_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["client_id"],
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				planModifiers.IgnoreCase(),
			},
		},
	}
}

func (r *res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.ValueString()) }).
		Then(func() { user = sql.CreateUser(ctx, db, req.Plan.toSettings()) }).
		Then(func() {
			req.Plan.Id = types.StringValue(common2.DbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String())
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
		Then(func() { id = common2.ParseDbObjectId[sql.UserId](ctx, req.State.Id.ValueString()) }).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, id.DbId) }).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() {
			state := req.State.withSettings(ctx, user.GetSettings(ctx))
			state.DatabaseId = types.StringValue(fmt.Sprint(id.DbId))
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
			db = common2.GetResourceDb(ctx, req.Conn, req.State.DatabaseId.ValueString())
			id = common2.ParseDbObjectId[sql.UserId](ctx, req.State.Id.ValueString())
		}).
		Then(func() { user = sql.GetUser(ctx, db, id.ObjectId) }).
		Then(func() { user.Drop(ctx) })
}
