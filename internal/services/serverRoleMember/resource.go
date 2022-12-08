package serverRoleMember

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"strings"
)

type resourceData struct {
	Id       types.String `tfsdk:"id"`
	RoleId   types.String `tfsdk:"role_id"`
	MemberId types.String `tfsdk:"member_id"`
}

type res struct{}

func (r res) GetName() string {
	return "server_role_member"
}

func (r res) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Manages server role membership."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "`<role_id>/<member_id>`. Role and member IDs can be retrieved using `mssql_server_role` or `mssql_sql_login`",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"role_id": schema.StringAttribute{
			MarkdownDescription: "ID of the server role. Can be retrieved using `mssql_server_role`",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"member_id": schema.StringAttribute{
			MarkdownDescription: "ID of the member. Can be retrieved using `mssql_server_role` or `mssql_sql_login`",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	}
}

func (r res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	roleId, memberId := r.parseInputs(ctx, req.State)
	var role sql.ServerRole

	req.
		Then(func() { role = sql.GetServerRole(ctx, req.Conn, roleId) }).
		Then(func() {
			req.State.RoleId = types.StringValue(fmt.Sprint(roleId))
			req.State.MemberId = types.StringValue(fmt.Sprint(memberId))

			if role.HasMember(ctx, memberId) {
				resp.SetState(req.State)
			}
		})
}

func (r res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	roleId, memberId := r.parseInputs(ctx, req.Plan)
	var role sql.ServerRole

	req.
		Then(func() { role = sql.GetServerRole(ctx, req.Conn, roleId) }).
		Then(func() { role.AddMember(ctx, memberId) }).
		Then(func() {
			resp.State = req.Plan
			resp.State.Id = types.StringValue(fmt.Sprintf("%d/%d", roleId, memberId))
		})
}

func (r res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	panic("not supported")
}

func (r res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], resp *resource.DeleteResponse[resourceData]) {
	roleId, memberId := r.parseInputs(ctx, req.State)
	var role sql.ServerRole

	req.
		Then(func() { role = sql.GetServerRole(ctx, req.Conn, roleId) }).
		Then(func() { role.RemoveMember(ctx, memberId) })
}

func (r res) parseInputs(ctx context.Context, data resourceData) (sql.ServerRoleId, sql.GenericServerPrincipalId) {
	parseId := func(idStr string) int {
		id, err := strconv.Atoi(idStr)
		utils.AddError(ctx, "Failed to parse ID", err)
		return id
	}

	if common.IsAttrSet(data.Id) {
		parts := strings.Split(data.Id.ValueString(), "/")

		if len(parts) != 2 {
			utils.AddError(ctx, "Invalid ID format", fmt.Errorf("expected 2 parts of ID, got %d", len(parts)))
			return 0, 0
		}

		return sql.ServerRoleId(parseId(parts[0])), sql.GenericServerPrincipalId(parseId(parts[1]))
	}

	return sql.ServerRoleId(parseId(data.RoleId.ValueString())), sql.GenericServerPrincipalId(parseId(data.MemberId.ValueString()))
}
