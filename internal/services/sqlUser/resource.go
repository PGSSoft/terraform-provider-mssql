package sqlUser

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"strconv"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
)

type res struct{}

func (r *res) GetName() string {
	return "sql_user"
}

func (r *res) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Manages database-level user, based on SQL login."
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
		"database_id": schema.StringAttribute{
			MarkdownDescription: common2.AttributeDescriptions["database_id"] + " Defaults to ID of `master`.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"login_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["login_id"],
			Required:            true,
		},
	}
}

func (r *res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.ValueString()) }).
		Then(func() { user = sql.CreateUser(ctx, db, req.Plan.toSettings()) }).
		Then(func() { resp.State = req.Plan.withIds(db.GetId(ctx), user.GetId(ctx)) })
}

func (r *res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var user sql.User

	req.
		Then(func() { user = getUser(ctx, req.Conn, req.State) }).
		Then(func() {
			state := req.State.withIds(user.GetDatabaseId(ctx), user.GetId(ctx))
			resp.SetState(state.withSettings(user.GetSettings(ctx)))
		})
}

func (r *res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	var user sql.User

	req.
		Then(func() { user = getUser(ctx, req.Conn, req.Plan) }).
		Then(func() { user.UpdateSettings(ctx, req.Plan.toSettings()) }).
		Then(func() { resp.State = req.Plan.withSettings(user.GetSettings(ctx)) })
}

func (r *res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], resp *resource.DeleteResponse[resourceData]) {
	var user sql.User

	req.
		Then(func() { user = getUser(ctx, req.Conn, req.State) }).
		Then(func() { user.Drop(ctx) })
}

func getUser(ctx context.Context, conn sql.Connection, data resourceData) sql.User {
	idSegments := strings.Split(data.Id.ValueString(), "/")
	id, err := strconv.Atoi(idSegments[1])
	if err != nil {
		utils.AddError(ctx, "Error converting user ID", err)
		return nil
	}

	db := common2.GetResourceDb(ctx, conn, idSegments[0])
	if utils.HasError(ctx) {
		return nil
	}

	return sql.GetUser(ctx, db, sql.UserId(id))
}
