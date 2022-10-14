package databaseRole

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type res struct{}

func (r *res) GetName() string {
	return "database_role"
}

func (r *res) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		Description: "Manages database-level role.",
		Attributes: map[string]tfsdk.Attribute{
			"id":          common2.ToResourceId(roleAttributes["id"]),
			"name":        common2.ToRequired(roleAttributes["name"]),
			"database_id": common2.DatabaseIdResourceAttribute,
			"owner_id": func() tfsdk.Attribute {
				attr := roleAttributes["owner_id"]
				attr.Optional = true
				attr.Computed = true
				attr.MarkdownDescription += " Defaults to ID of current user, used to authorize the Terraform provider."
				return attr
			}(),
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
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() {
			if !req.Plan.OwnerId.Null && !req.Plan.OwnerId.Unknown {
				ownerId = common2.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.OwnerId.Value)

				if ownerId.DbId != dbId {
					utils.AddError(ctx, "Role owner must be principal defined in the same DB as the role", errors.New("owner and principal DBs are different"))
				}
			}
		}).
		Then(func() {
			if ownerId.IsEmpty {
				role = sql.CreateDatabaseRole(ctx, db, req.Plan.Name.Value, sql.EmptyDatabasePrincipalId)
			} else {
				role = sql.CreateDatabaseRole(ctx, db, req.Plan.Name.Value, ownerId.ObjectId)
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
		Then(func() { roleId = common2.ParseDbObjectId[sql.DatabaseRoleId](ctx, req.State.Id.Value) }).
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
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() { roleId = common2.ParseDbObjectId[sql.DatabaseRoleId](ctx, req.Plan.Id.Value) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() {
			if !req.Plan.OwnerId.Null && !req.Plan.OwnerId.Unknown {
				ownerId = common2.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.OwnerId.Value)

				if ownerId.DbId != dbId {
					utils.AddError(ctx, "Role owner must be principal defined in the same DB as the role", errors.New("owner and principal DBs are different"))
				}
			}
		}).
		Then(func() {
			if role.GetName(ctx) != req.Plan.Name.Value && !utils.HasError(ctx) {
				role.Rename(ctx, req.Plan.Name.Value)
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
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.State.DatabaseId.Value) }).
		Then(func() { roleId = common2.ParseDbObjectId[sql.DatabaseRoleId](ctx, req.State.Id.Value) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { role.Drop(ctx) })
}
