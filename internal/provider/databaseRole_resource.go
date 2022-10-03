package provider

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/resource"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type databaseRoleResource struct {
	BaseResource
}

func (p mssqlProvider) NewDatabaseRoleResource() func() sdkresource.Resource {
	return func() sdkresource.Resource {
		return resource.WrapResource[databaseRoleResourceData](&databaseRoleResource{})
	}
}

func (r *databaseRoleResource) GetName() string {
	return "database_role"
}

func (r *databaseRoleResource) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		Description: "Manages database-level role.",
		Attributes: map[string]tfsdk.Attribute{
			"id":          toResourceId(databaseRoleAttributes["id"]),
			"name":        toRequired(databaseRoleAttributes["name"]),
			"database_id": databaseIdResourceAttribute,
			"owner_id": func() tfsdk.Attribute {
				attr := databaseRoleAttributes["owner_id"]
				attr.Optional = true
				attr.Computed = true
				attr.MarkdownDescription += " Defaults to ID of current user, used to authorize the Terraform provider."
				return attr
			}(),
		},
	}
}

func (r *databaseRoleResource) Create(ctx context.Context, req resource.CreateRequest[databaseRoleResourceData], resp *resource.CreateResponse[databaseRoleResourceData]) {
	var (
		db   sql.Database
		dbId sql.DatabaseId
		role sql.DatabaseRole
	)
	ownerId := dbObjectId[sql.GenericDatabasePrincipalId]{IsEmpty: true}

	req.
		Then(func() { db = getResourceDb(ctx, r.conn, req.Plan.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() {
			if !req.Plan.OwnerId.Null && !req.Plan.OwnerId.Unknown {
				ownerId = parseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.OwnerId.Value)

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

func (r *databaseRoleResource) Read(ctx context.Context, req resource.ReadRequest[databaseRoleResourceData], resp *resource.ReadResponse[databaseRoleResourceData]) {
	var (
		db     sql.Database
		roleId dbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)

	req.
		Then(func() { roleId = parseDbObjectId[sql.DatabaseRoleId](ctx, req.State.Id.Value) }).
		Then(func() { db = sql.GetDatabase(ctx, r.conn, roleId.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { resp.SetState(req.State.withRoleData(ctx, role)) })
}

func (r *databaseRoleResource) Update(ctx context.Context, req resource.UpdateRequest[databaseRoleResourceData], resp *resource.UpdateResponse[databaseRoleResourceData]) {
	var (
		dbId   sql.DatabaseId
		db     sql.Database
		roleId dbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)
	ownerId := dbObjectId[sql.GenericDatabasePrincipalId]{IsEmpty: true}

	req.
		Then(func() { db = getResourceDb(ctx, r.conn, req.Plan.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() { roleId = parseDbObjectId[sql.DatabaseRoleId](ctx, req.Plan.Id.Value) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() {
			if !req.Plan.OwnerId.Null && !req.Plan.OwnerId.Unknown {
				ownerId = parseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.OwnerId.Value)

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

func (r *databaseRoleResource) Delete(ctx context.Context, req resource.DeleteRequest[databaseRoleResourceData], _ *resource.DeleteResponse[databaseRoleResourceData]) {
	var (
		db     sql.Database
		roleId dbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)

	req.
		Then(func() { db = getResourceDb(ctx, r.conn, req.State.DatabaseId.Value) }).
		Then(func() { roleId = parseDbObjectId[sql.DatabaseRoleId](ctx, req.State.Id.Value) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { role.Drop(ctx) })
}
