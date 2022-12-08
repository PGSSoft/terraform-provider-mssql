package databasePermission

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type resourceData struct {
	Id              types.String `tfsdk:"id"`
	PrincipalId     types.String `tfsdk:"principal_id"`
	Permission      types.String `tfsdk:"permission"`
	WithGrantOption types.Bool   `tfsdk:"with_grant_option"`
}

func (d resourceData) withPermission(perm sql.DatabasePermission) resourceData {
	d.Permission = types.StringValue(perm.Name)
	d.WithGrantOption = types.BoolValue(perm.WithGrantOption)
	return d
}

func (d resourceData) toPermission() sql.DatabasePermission {
	return sql.DatabasePermission{
		Name:            d.Permission.ValueString(),
		WithGrantOption: d.WithGrantOption.ValueBool(),
	}
}

type res struct{}

func (r res) GetName() string {
	return "database_permission"
}

func (r res) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Grants database-level permission."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["id"],
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"principal_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["principal_id"],
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"permission": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["permission"],
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"with_grant_option": schema.BoolAttribute{
			MarkdownDescription: attrDescriptions["with_grant_option"] + " Defaults to `false`.",
			Optional:            true,
			Computed:            true,
		},
	}
}

func (r res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var permissions sql.DatabasePermissions
	db, principalId, permission := r.parseInputs(ctx, req.Conn, req.State)

	req.
		Then(func() { permissions = db.GetPermissions(ctx, principalId.ObjectId) }).
		Then(func() {
			req.State.PrincipalId = types.StringValue(principalId.String())

			if perm, ok := permissions[permission]; ok {
				resp.SetState(req.State.withPermission(perm))
			}
		})
}

func (r res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var (
		db          sql.Database
		principalId common.DbObjectId[sql.GenericDatabasePrincipalId]
	)

	req.
		Then(func() {
			principalId = common.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.PrincipalId.ValueString())
		}).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, principalId.DbId) }).
		Then(func() { db.GrantPermission(ctx, principalId.ObjectId, req.Plan.toPermission()) }).
		Then(func() {
			req.Plan.WithGrantOption = types.BoolValue(req.Plan.WithGrantOption.ValueBool())
			req.Plan.Id = types.StringValue(fmt.Sprintf("%v/%s", principalId, req.Plan.Permission.ValueString()))

			resp.State = req.Plan
		})
}

func (r res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	db, principalId, _ := r.parseInputs(ctx, req.Conn, req.Plan)

	req.
		Then(func() { db.UpdatePermission(ctx, principalId.ObjectId, req.Plan.toPermission()) }).
		Then(func() { resp.State = req.Plan.withPermission(req.Plan.toPermission()) })
}

func (r res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	db, principalId, permission := r.parseInputs(ctx, req.Conn, req.State)

	req.Then(func() { db.RevokePermission(ctx, principalId.ObjectId, permission) })
}

func (r res) parseInputs(ctx context.Context, conn sql.Connection, data resourceData) (sql.Database, common.DbObjectId[sql.GenericDatabasePrincipalId], string) {
	parts := strings.Split(data.Id.ValueString(), "/")
	permission := parts[len(parts)-1]
	principalId := common.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, data.Id.ValueString()[:len(data.Id.ValueString())-len(permission)-1])

	return sql.GetDatabase(ctx, conn, principalId.DbId), principalId, permission
}
