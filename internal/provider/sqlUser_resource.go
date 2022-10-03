package provider

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/resource"
	"strconv"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type sqlUserResource struct {
	BaseResource
}

func (p mssqlProvider) NewSqlUserResource() func() sdkresource.Resource {
	return func() sdkresource.Resource {
		return resource.WrapResource[sqlUserResourceData](&sqlUserResource{})
	}
}

func (r *sqlUserResource) GetName() string {
	return "sql_user"
}

func (r *sqlUserResource) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		Description: "Manages database-level user, based on SQL login.",
		Attributes: map[string]tfsdk.Attribute{
			"id":          toResourceId(sqlUserAttributes["id"]),
			"name":        toRequired(sqlUserAttributes["name"]),
			"database_id": databaseIdResourceAttribute,
			"login_id":    toRequired(sqlUserAttributes["login_id"]),
		},
	}
}

func (r *sqlUserResource) Create(ctx context.Context, req resource.CreateRequest[sqlUserResourceData], resp *resource.CreateResponse[sqlUserResourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { db = getResourceDb(ctx, r.conn, req.Plan.DatabaseId.Value) }).
		Then(func() { user = sql.CreateUser(ctx, db, req.Plan.toSettings()) }).
		Then(func() { resp.State = req.Plan.withIds(db.GetId(ctx), user.GetId(ctx)) })
}

func (r *sqlUserResource) Read(ctx context.Context, req resource.ReadRequest[sqlUserResourceData], resp *resource.ReadResponse[sqlUserResourceData]) {
	var user sql.User

	req.
		Then(func() { user = r.getUser(ctx, req.State) }).
		Then(func() {
			state := req.State.withIds(user.GetDatabaseId(ctx), user.GetId(ctx))
			resp.SetState(state.withSettings(user.GetSettings(ctx)))
		})
}

func (r sqlUserResource) Update(ctx context.Context, req resource.UpdateRequest[sqlUserResourceData], resp *resource.UpdateResponse[sqlUserResourceData]) {
	var user sql.User

	req.
		Then(func() { user = r.getUser(ctx, req.Plan) }).
		Then(func() { user.UpdateSettings(ctx, req.Plan.toSettings()) }).
		Then(func() { resp.State = req.Plan.withSettings(user.GetSettings(ctx)) })
}

func (r sqlUserResource) Delete(ctx context.Context, req resource.DeleteRequest[sqlUserResourceData], resp *resource.DeleteResponse[sqlUserResourceData]) {
	var user sql.User

	req.
		Then(func() { user = r.getUser(ctx, req.State) }).
		Then(func() { user.Drop(ctx) })
}

func (r sqlUserResource) getUser(ctx context.Context, data sqlUserResourceData) sql.User {
	idSegments := strings.Split(data.Id.Value, "/")
	id, err := strconv.Atoi(idSegments[1])
	if err != nil {
		utils.AddError(ctx, "Error converting user ID", err)
		return nil
	}

	db := getResourceDb(ctx, r.conn, idSegments[0])
	if utils.HasError(ctx) {
		return nil
	}

	return sql.GetUser(ctx, db, sql.UserId(id))
}
