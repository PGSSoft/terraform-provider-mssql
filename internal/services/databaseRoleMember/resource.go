package databaseRoleMember

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type res struct{}

func (r *res) GetName() string {
	return "database_role_member"
}

func (r *res) GetSchema(context.Context) tfsdk.Schema {
	annotatePrincipalId := func(attrName string) tfsdk.Attribute {
		attr := attributes[attrName]
		attr.Required = true
		attr.PlanModifiers = tfsdk.AttributePlanModifiers{sdkresource.RequiresReplace()}
		return attr
	}

	return tfsdk.Schema{
		Description: "Manages database role membership.",
		Attributes: map[string]tfsdk.Attribute{
			"id":        common2.ToResourceId(attributes["id"]),
			"role_id":   annotatePrincipalId("role_id"),
			"member_id": annotatePrincipalId("member_id"),
		},
	}
}

func (r *res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var (
		roleId   common2.DbObjectId[sql.DatabaseRoleId]
		memberId common2.DbObjectId[sql.GenericDatabasePrincipalId]
		db       sql.Database
		role     sql.DatabaseRole
	)

	req.
		Then(func() {
			roleId = common2.ParseDbObjectId[sql.DatabaseRoleId](ctx, req.Plan.RoleId.Value)
			memberId = common2.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.MemberId.Value)
		}).
		Then(func() {
			if roleId.DbId != memberId.DbId {
				utils.AddError(ctx, "Role and member must be defined in the same database", errors.New("DB IDs of role and member are different"))
			}
		}).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, roleId.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { role.AddMember(ctx, memberId.ObjectId) }).
		Then(func() {
			req.Plan.Id = types.String{Value: common2.DbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]{DbObjectId: roleId, MemberId: memberId.ObjectId}.String()}
			resp.State = req.Plan
		})
}

func (r *res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var (
		id   common2.DbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]
		db   sql.Database
		role sql.DatabaseRole
	)

	req.
		Then(func() {
			id = common2.ParseDbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId](ctx, req.State.Id.Value)
		}).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, id.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, id.ObjectId) }).
		Then(func() {
			if role.HasMember(ctx, id.MemberId) {
				req.State.RoleId = types.String{Value: id.DbObjectId.String()}
				req.State.MemberId = types.String{Value: id.GetMemberId().String()}
				resp.SetState(req.State)
			}
		})
}

func (r *res) Update(context.Context, resource.UpdateRequest[resourceData], *resource.UpdateResponse[resourceData]) {
	panic("Resource does not support updates. All changes should trigger recreate.")
}

func (r *res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	var (
		id   common2.DbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]
		db   sql.Database
		role sql.DatabaseRole
	)

	req.
		Then(func() {
			id = common2.ParseDbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId](ctx, req.State.Id.Value)
		}).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, id.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, id.ObjectId) }).
		Then(func() { role.RemoveMember(ctx, id.MemberId) })
}
