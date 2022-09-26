package provider

import (
	"context"
	"strconv"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &sqlUserResource{}
	_ resource.ResourceWithImportState = sqlUserResource{}
)

type sqlUserResource struct {
	sqlUserResourceBase
}

func (p mssqlProvider) NewSqlUserResource() func() resource.Resource {
	return func() resource.Resource {
		return &sqlUserResource{}
	}
}

func (s sqlUserResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "mssql_sql_user"
}

func (r *sqlUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (rt sqlUserResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Manages database-level user, based on SQL login.",
		Attributes: map[string]tfsdk.Attribute{
			"id":          toResourceId(sqlUserAttributes["id"]),
			"name":        toRequired(sqlUserAttributes["name"]),
			"database_id": databaseIdResourceAttribute,
			"login_id":    toRequired(sqlUserAttributes["login_id"]),
		},
	}, nil
}

func (r sqlUserResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlUserResourceData](ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	db := getResourceDb(ctx, r.Db, data.DatabaseId.Value)
	if utils.HasError(ctx) {
		return
	}

	user := sql.CreateUser(ctx, db, data.toSettings())
	if utils.HasError(ctx) {
		return
	}

	data = data.withIds(db.GetId(ctx), user.GetId(ctx))
	if utils.HasError(ctx) {
		return
	}

	utils.SetData(ctx, &response.State, data)
}

func (r sqlUserResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlUserResourceData](ctx, request.State)
	if utils.HasError(ctx) {
		return
	}

	user := r.getUser(ctx, data)
	if utils.HasError(ctx) {
		return
	}

	data = data.withIds(user.GetDatabaseId(ctx), user.GetId(ctx)).withSettings(user.GetSettings(ctx))
	if utils.HasError(ctx) {
		return
	}

	utils.SetData(ctx, &response.State, data)
}

func (r sqlUserResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlUserResourceData](ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	user := r.getUser(ctx, data)
	if utils.HasError(ctx) {
		return
	}

	user.UpdateSettings(ctx, data.toSettings())

	utils.SetData(ctx, &response.State, data.withSettings(user.GetSettings(ctx)))
}

func (r sqlUserResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlUserResourceData](ctx, request.State)
	if utils.HasError(ctx) {
		return
	}

	user := r.getUser(ctx, data)
	if utils.HasError(ctx) {
		return
	}

	if user.Drop(ctx); utils.HasError(ctx) {
		return
	}

	response.State.RemoveResource(ctx)
}

func (r sqlUserResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r sqlUserResource) getUser(ctx context.Context, data sqlUserResourceData) sql.User {
	idSegments := strings.Split(data.Id.Value, "/")
	id, err := strconv.Atoi(idSegments[1])
	if err != nil {
		utils.AddError(ctx, "Error converting user ID", err)
		return nil
	}

	db := getResourceDb(ctx, r.Db, idSegments[0])
	if utils.HasError(ctx) {
		return nil
	}

	return sql.GetUser(ctx, db, sql.UserId(id))
}
