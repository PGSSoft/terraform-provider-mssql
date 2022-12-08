package databaseRoleMember

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceData struct {
	Id       types.String `tfsdk:"id"`
	RoleId   types.String `tfsdk:"role_id"`
	MemberId types.String `tfsdk:"member_id"`
}

type res struct{}

func (r *res) GetName() string {
	return "database_role_member"
}

func (r *res) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Manages database role membership."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "`<database_id>/<role_id>/<member_id>`. Role and member IDs can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<name>')`",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"role_id": schema.StringAttribute{
			MarkdownDescription: "`<database_id>/<role_id>`",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"member_id": schema.StringAttribute{
			MarkdownDescription: "Can be either user or role ID in format `<database_id>/<member_id>`. Can be retrieved using `mssql_sql_user` or `mssql_database_member`.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
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
			roleId = common2.ParseDbObjectId[sql.DatabaseRoleId](ctx, req.Plan.RoleId.ValueString())
			memberId = common2.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.MemberId.ValueString())
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
			req.Plan.Id = types.StringValue(common2.DbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId]{DbObjectId: roleId, MemberId: memberId.ObjectId}.String())
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
			id = common2.ParseDbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId](ctx, req.State.Id.ValueString())
		}).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, id.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, id.ObjectId) }).
		Then(func() {
			if role.HasMember(ctx, id.MemberId) {
				req.State.RoleId = types.StringValue(id.DbObjectId.String())
				req.State.MemberId = types.StringValue(id.GetMemberId().String())
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
			id = common2.ParseDbObjectMemberId[sql.DatabaseRoleId, sql.GenericDatabasePrincipalId](ctx, req.State.Id.ValueString())
		}).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, id.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, id.ObjectId) }).
		Then(func() { role.RemoveMember(ctx, id.MemberId) })
}
