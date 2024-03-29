package databaseRole

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
)

type res struct{}

func (r *res) GetName() string {
	return "database_role"
}

func (r *res) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Manages database-level role."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: roleAttributeDescriptions["id"],
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: roleAttributeDescriptions["name"],
			Required:            true,
			Validators:          validators.UserNameValidators,
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common2.AttributeDescriptions["database_id"] + " Defaults to ID of `master`.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"owner_id": schema.StringAttribute{
			MarkdownDescription: roleAttributeDescriptions["owner_id"] + " Defaults to ID of current user, used to authorize the Terraform provider.",
			Optional:            true,
			Computed:            true,
		},
	}
}

func (r *res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var (
		db   sql.Database
		dbId sql.DatabaseId
		role sql.DatabaseRole
	)
	ownerId := common2.DbObjectId[sql.GenericDatabasePrincipalId]{IsEmpty: true}

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.ValueString()) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() {
			if common2.IsAttrSet(req.Plan.OwnerId) {
				ownerId = common2.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.OwnerId.ValueString())

				if ownerId.DbId != dbId {
					utils.AddError(ctx, "Role owner must be principal defined in the same DB as the role", errors.New("owner and principal DBs are different"))
				}
			}
		}).
		Then(func() {
			if ownerId.IsEmpty {
				role = sql.CreateDatabaseRole(ctx, db, req.Plan.Name.ValueString(), sql.EmptyDatabasePrincipalId)
			} else {
				role = sql.CreateDatabaseRole(ctx, db, req.Plan.Name.ValueString(), ownerId.ObjectId)
			}
		}).
		Then(func() { resp.State = req.Plan.withRoleData(ctx, role) })
}

func (r *res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var (
		db     sql.Database
		roleId common2.DbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)

	req.
		Then(func() { roleId = common2.ParseDbObjectId[sql.DatabaseRoleId](ctx, req.State.Id.ValueString()) }).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, roleId.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { resp.SetState(req.State.withRoleData(ctx, role)) })
}

func (r *res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	var (
		dbId   sql.DatabaseId
		db     sql.Database
		roleId common2.DbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)
	ownerId := common2.DbObjectId[sql.GenericDatabasePrincipalId]{IsEmpty: true}

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.ValueString()) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() { roleId = common2.ParseDbObjectId[sql.DatabaseRoleId](ctx, req.Plan.Id.ValueString()) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() {
			if common2.IsAttrSet(req.Plan.OwnerId) {
				ownerId = common2.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.OwnerId.ValueString())

				if ownerId.DbId != dbId {
					utils.AddError(ctx, "Role owner must be principal defined in the same DB as the role", errors.New("owner and principal DBs are different"))
				}
			}
		}).
		Then(func() {
			if role.GetName(ctx) != req.Plan.Name.ValueString() && !utils.HasError(ctx) {
				role.Rename(ctx, req.Plan.Name.ValueString())
			}
		}).
		Then(func() {
			if role.GetOwnerId(ctx) != ownerId.ObjectId && !utils.HasError(ctx) {
				if ownerId.IsEmpty {
					role.ChangeOwner(ctx, sql.EmptyDatabasePrincipalId)
				} else {
					role.ChangeOwner(ctx, ownerId.ObjectId)
				}
			}

			resp.State = req.Plan
		})
}

func (r *res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	var (
		db     sql.Database
		roleId common2.DbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.State.DatabaseId.ValueString()) }).
		Then(func() { roleId = common2.ParseDbObjectId[sql.DatabaseRoleId](ctx, req.State.Id.ValueString()) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { role.Drop(ctx) })
}
