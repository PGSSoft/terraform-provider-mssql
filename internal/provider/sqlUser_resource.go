package provider

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strconv"
	"strings"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.ResourceType            = SqlUserResourceType{}
	_ tfsdk.Resource                = sqlUserResource{}
	_ tfsdk.ResourceWithImportState = sqlUserResource{}
)

type SqlUserResourceType struct{}

func (rt SqlUserResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Manages database-level user, based on SQL login.",
		Attributes: map[string]tfsdk.Attribute{
			"id": func() tfsdk.Attribute {
				attr := sqlUserAttributes["id"]
				attr.Computed = true
				attr.PlanModifiers = tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				}
				return attr
			}(),
			"name": func() tfsdk.Attribute {
				attr := sqlUserAttributes["name"]
				attr.Required = true
				return attr
			}(),
			"database_id": func() tfsdk.Attribute {
				attr := sqlUserAttributes["database_id"]
				attr.Optional = true
				attr.Computed = true
				attr.MarkdownDescription += " Defaults to ID of `master`."
				attr.PlanModifiers = tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				}
				return attr
			}(),
			"login_id": func() tfsdk.Attribute {
				attr := sqlUserAttributes["login_id"]
				attr.Required = true
				return attr
			}(),
		},
	}, nil
}

func (rt SqlUserResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) sqlUserResource {
		return sqlUserResource{sqlUserResourceBase: sqlUserResourceBase{Resource: base}}
	})
}

type sqlUserResource struct {
	sqlUserResourceBase
}

func (r sqlUserResource) Create(ctx context.Context, request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlUserResourceData](ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	db := r.getDb(ctx, data.DatabaseId.Value)
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

func (r sqlUserResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {
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

func (r sqlUserResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest, response *tfsdk.UpdateResourceResponse) {
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

func (r sqlUserResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
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

func (r sqlUserResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), request, response)
}

func (r sqlUserResource) getUser(ctx context.Context, data sqlUserResourceData) sql.User {
	idSegments := strings.Split(data.Id.Value, "/")
	id, err := strconv.Atoi(idSegments[1])
	if err != nil {
		utils.AddError(ctx, "Error converting user ID", err)
		return nil
	}

	db := r.getDb(ctx, idSegments[0])
	if utils.HasError(ctx) {
		return nil
	}

	return sql.GetUser(ctx, db, sql.UserId(id))
}
