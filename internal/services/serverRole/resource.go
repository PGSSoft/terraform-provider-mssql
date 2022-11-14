package serverRole

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type res struct{}

func (r res) GetName() string {
	return "server_role"
}

func (r res) GetSchema(ctx context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: "Manages server-level role.",
		Attributes: map[string]tfsdk.Attribute{
			"id":   common.ToResourceId(attributes["id"]),
			"name": common.ToRequired(attributes["name"]),
			"owner_id": func() tfsdk.Attribute {
				attr := attributes["owner_id"]
				attr.Optional = true
				attr.Computed = true
				attr.PlanModifiers = tfsdk.AttributePlanModifiers{
					sdkresource.RequiresReplace(),
				}
				return attr
			}(),
		},
	}
}

func (r res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	id := parseId(ctx, req.State.Id)
	var role sql.ServerRole

	req.
		Then(func() { role = sql.GetServerRole(ctx, req.Conn, id) }).
		Then(func() { resp.SetState(req.State.withSettings(role.GetSettings(ctx))) })
}

func (r res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	settings := req.Plan.toSettings(ctx)

	var role sql.ServerRole

	req.
		Then(func() { role = sql.CreateServerRole(ctx, req.Conn, settings) }).
		Then(func() {
			resp.State = req.Plan.withSettings(role.GetSettings(ctx))
			resp.State.Id = types.StringValue(fmt.Sprint(role.GetId(ctx)))
		})
}

func (r res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	id := parseId(ctx, req.Plan.Id)
	var role sql.ServerRole

	req.
		Then(func() { role = sql.GetServerRole(ctx, req.Conn, id) }).
		Then(func() { role.Rename(ctx, req.Plan.Name.ValueString()) }).
		Then(func() { resp.State = req.Plan.withSettings(role.GetSettings(ctx)) })
}

func (r res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], resp *resource.DeleteResponse[resourceData]) {
	id := parseId(ctx, req.State.Id)
	var role sql.ServerRole

	req.
		Then(func() { role = sql.GetServerRole(ctx, req.Conn, id) }).
		Then(func() { role.Drop(ctx) })
}
