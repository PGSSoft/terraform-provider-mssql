package serverPermission

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/attrs"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceData struct {
	Id              attrs.PermissionId[sql.GenericServerPrincipalId] `tfsdk:"id"`
	PrincipalId     attrs.NumericId[sql.GenericServerPrincipalId]    `tfsdk:"principal_id"`
	Permission      types.String                                     `tfsdk:"permission"`
	WithGrantOption types.Bool                                       `tfsdk:"with_grant_option"`
}

type res struct{}

func (r res) GetName() string {
	return "server_permission"
}

func (r res) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: "Grants server-level permission.",
		Attributes: map[string]tfsdk.Attribute{
			"id":           common.ToResourceId(attributes["id"]),
			"principal_id": common.ToRequiredImmutable(attributes["principal_id"]),
			"permission":   common.ToRequiredImmutable(attributes["permission"]),
			"with_grant_option": func() tfsdk.Attribute {
				attr := attributes["with_grant_option"]
				attr.MarkdownDescription += " Defaults to `false`"
				attr.Optional = true
				attr.Computed = true

				return attr
			}(),
		},
	}
}

func (r res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var perms sql.ServerPermissions
	principalId := req.State.Id.ObjectId(ctx)

	req.
		Then(func() { perms = req.Conn.GetPermissions(ctx, principalId) }).
		Then(func() {
			if perm, ok := perms[req.State.Id.Permission()]; ok {
				req.State.Permission = types.StringValue(perm.Name)
				req.State.PrincipalId = attrs.NumericIdValue(principalId)
				req.State.WithGrantOption = types.BoolValue(perm.WithGrantOption)
				resp.SetState(req.State)
			}
		})
}

func (r res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var pId sql.GenericServerPrincipalId
	req.
		Then(func() { pId = req.Plan.PrincipalId.Id(ctx) }).
		Then(func() {
			req.Conn.GrantPermission(ctx, pId, sql.ServerPermission{Name: req.Plan.Permission.ValueString(), WithGrantOption: req.Plan.WithGrantOption.ValueBool()})
		}).
		Then(func() {
			resp.State = req.Plan
			resp.State.Id = attrs.PermissionIdValue(pId, req.Plan.Permission.ValueString())
			resp.State.WithGrantOption = types.BoolValue(req.Plan.WithGrantOption.ValueBool())
		})
}

func (r res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	principalId := req.State.Id.ObjectId(ctx)

	req.
		Then(func() {
			req.Conn.GrantPermission(ctx, principalId, sql.ServerPermission{Name: req.Plan.Permission.ValueString(), WithGrantOption: req.Plan.WithGrantOption.ValueBool()})
		}).
		Then(func() { resp.State = req.Plan })
}

func (r res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	principalId := req.State.Id.ObjectId(ctx)

	req.Then(func() { req.Conn.RevokePermission(ctx, principalId, req.State.Permission.ValueString()) })
}
