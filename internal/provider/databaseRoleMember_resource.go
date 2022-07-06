package provider

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.ResourceType            = DatabaseRoleMemberResourceType{}
	_ tfsdk.Resource                = databaseRoleMemberResource{}
	_ tfsdk.ResourceWithImportState = databaseRoleMemberResource{}
)

type DatabaseRoleMemberResourceType struct{}

func (d DatabaseRoleMemberResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	annotatePrincipalId := func(attrName string) tfsdk.Attribute {
		attr := databaseRoleMemberAttributes[attrName]
		attr.Required = true
		attr.PlanModifiers = tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()}
		return attr
	}

	return tfsdk.Schema{
		Description: "Manages database role membership.",
		Attributes: map[string]tfsdk.Attribute{
			"id":        toResourceId(databaseRoleMemberAttributes["id"]),
			"role_id":   annotatePrincipalId("role_id"),
			"member_id": annotatePrincipalId("member_id"),
		},
	}, nil
}

func (d DatabaseRoleMemberResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) databaseRoleMemberResource {
		return databaseRoleMemberResource{Resource: base}
	})
}

type databaseRoleMemberResource struct {
	Resource
}

func (d databaseRoleMemberResource) Create(ctx context.Context, request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	var (
		data     databaseRoleMemberResourceData
		roleId   dbObjectId[sql.DatabaseRoleId]
		memberId dbObjectId[sql.GenericDatabasePrincipalId]
		db       sql.Database
		role     sql.DatabaseRole
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleMemberResourceData](ctx, request.Plan) }).
		Then(func() {
			roleId = parseDbObjectId[sql.DatabaseRoleId](ctx, data.RoleId.Value)
			memberId = parseDbObjectId[sql.GenericDatabasePrincipalId](ctx, data.MemberId.Value)
		}).
		Then(func() {
			if roleId.DbId != memberId.DbId {
				utils.AddError(ctx, "Role and member must be defined in the same database", errors.New("DB IDs of role and member are different"))
			}
		}).
		Then(func() { db = sql.GetDatabase(ctx, d.Db, roleId.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { role.AddMember(ctx, memberId.ObjectId) }).
		Then(func() {
			data.Id = types.String{Value: dbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]{dbObjectId: roleId, MemberId: memberId.ObjectId}.String()}
			utils.SetData(ctx, &response.State, data)
		})
}

func (d databaseRoleMemberResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {
	var (
		data databaseRoleMemberResourceData
		id   dbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]
		db   sql.Database
		role sql.DatabaseRole
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleMemberResourceData](ctx, request.State) }).
		Then(func() {
			id = parseDbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId](ctx, data.Id.Value)
		}).
		Then(func() { db = sql.GetDatabase(ctx, d.Db, id.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, id.ObjectId) }).
		Then(func() {
			if !role.HasMember(ctx, id.MemberId) && !utils.HasError(ctx) {
				response.State.RemoveResource(ctx)
			} else {
				data.RoleId = types.String{Value: id.dbObjectId.String()}
				data.MemberId = types.String{Value: id.getMemberId().String()}
				utils.SetData(ctx, &response.State, data)
			}
		})
}

func (d databaseRoleMemberResource) Update(context.Context, tfsdk.UpdateResourceRequest, *tfsdk.UpdateResourceResponse) {
	panic("Resource does not support updates. All changes should trigger recreate.")
}

func (d databaseRoleMemberResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
	var (
		data databaseRoleMemberResourceData
		id   dbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]
		db   sql.Database
		role sql.DatabaseRole
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleMemberResourceData](ctx, request.State) }).
		Then(func() {
			id = parseDbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId](ctx, data.Id.Value)
		}).
		Then(func() { db = sql.GetDatabase(ctx, d.Db, id.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, id.ObjectId) }).
		Then(func() { role.RemoveMember(ctx, id.MemberId) }).
		Then(func() { response.State.RemoveResource(ctx) })
}

func (d databaseRoleMemberResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), request, response)
}
