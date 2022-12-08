package serverRole

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type res struct{}

func (r res) GetName() string {
	return "server_role"
}

func (r res) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Manages server-level role."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["id"],
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["name"],
			Required:            true,
			Validators:          validators.UserNameValidators,
		},
		"owner_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["owner_id"],
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplaceIfConfigured(),
			},
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
