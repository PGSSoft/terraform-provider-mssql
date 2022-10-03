package provider

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/resource"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type databaseRoleMemberResource struct {
	BaseResource
}

func (p mssqlProvider) NewDatabaseRoleMemberResource() func() sdkresource.Resource {
	return func() sdkresource.Resource {
		return resource.WrapResource[databaseRoleMemberResourceData](&databaseRoleMemberResource{})
	}
}

func (r *databaseRoleMemberResource) GetName() string {
	return "database_role_member"
}

func (r *databaseRoleMemberResource) GetSchema(context.Context) tfsdk.Schema {
	annotatePrincipalId := func(attrName string) tfsdk.Attribute {
		attr := databaseRoleMemberAttributes[attrName]
		attr.Required = true
		attr.PlanModifiers = tfsdk.AttributePlanModifiers{sdkresource.RequiresReplace()}
		return attr
	}

	return tfsdk.Schema{
		Description: "Manages database role membership.",
		Attributes: map[string]tfsdk.Attribute{
			"id":        toResourceId(databaseRoleMemberAttributes["id"]),
			"role_id":   annotatePrincipalId("role_id"),
			"member_id": annotatePrincipalId("member_id"),
		},
	}
}

func (r *databaseRoleMemberResource) Create(ctx context.Context, req resource.CreateRequest[databaseRoleMemberResourceData], resp *resource.CreateResponse[databaseRoleMemberResourceData]) {
	var (
		roleId   dbObjectId[sql.DatabaseRoleId]
		memberId dbObjectId[sql.GenericDatabasePrincipalId]
		db       sql.Database
		role     sql.DatabaseRole
	)

	req.
		Then(func() {
			roleId = parseDbObjectId[sql.DatabaseRoleId](ctx, req.Plan.RoleId.Value)
			memberId = parseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.MemberId.Value)
		}).
		Then(func() {
			if roleId.DbId != memberId.DbId {
				utils.AddError(ctx, "Role and member must be defined in the same database", errors.New("DB IDs of role and member are different"))
			}
		}).
		Then(func() { db = sql.GetDatabase(ctx, r.conn, roleId.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { role.AddMember(ctx, memberId.ObjectId) }).
		Then(func() {
			req.Plan.Id = types.String{Value: dbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]{dbObjectId: roleId, MemberId: memberId.ObjectId}.String()}
			resp.State = req.Plan
		})
}

func (r *databaseRoleMemberResource) Read(ctx context.Context, req resource.ReadRequest[databaseRoleMemberResourceData], resp *resource.ReadResponse[databaseRoleMemberResourceData]) {
	var (
		id   dbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]
		db   sql.Database
		role sql.DatabaseRole
	)

	req.
		Then(func() {
			id = parseDbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId](ctx, req.State.Id.Value)
		}).
		Then(func() { db = sql.GetDatabase(ctx, r.conn, id.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, id.ObjectId) }).
		Then(func() {
			if role.HasMember(ctx, id.MemberId) {
				req.State.RoleId = types.String{Value: id.dbObjectId.String()}
				req.State.MemberId = types.String{Value: id.getMemberId().String()}
				resp.SetState(req.State)
			}
		})
}

func (r *databaseRoleMemberResource) Update(context.Context, resource.UpdateRequest[databaseRoleMemberResourceData], *resource.UpdateResponse[databaseRoleMemberResourceData]) {
	panic("Resource does not support updates. All changes should trigger recreate.")
}

func (r *databaseRoleMemberResource) Delete(ctx context.Context, req resource.DeleteRequest[databaseRoleMemberResourceData], _ *resource.DeleteResponse[databaseRoleMemberResourceData]) {
	var (
		id   dbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]
		db   sql.Database
		role sql.DatabaseRole
	)

	req.
		Then(func() {
			id = parseDbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId](ctx, req.State.Id.Value)
		}).
		Then(func() { db = sql.GetDatabase(ctx, r.conn, id.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, id.ObjectId) }).
		Then(func() { role.RemoveMember(ctx, id.MemberId) })
}
