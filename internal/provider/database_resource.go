package provider

import (
	"context"
	"fmt"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &databaseResource{}
	_ resource.ResourceWithImportState = databaseResource{}
)

type databaseResource struct {
	Resource
}

func (p mssqlProvider) NewDatabaseResource() func() resource.Resource {
	return func() resource.Resource {
		return &databaseResource{}
	}
}

func (s databaseResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "mssql_database"
}

func (r *databaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (d databaseResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Manages single database.",
		Attributes: map[string]tfsdk.Attribute{
			"id":   toResourceId(databaseAttributes["id"]),
			"name": toRequired(databaseAttributes["name"]),
			"collation": func() tfsdk.Attribute {
				attr := databaseAttributes["collation"]
				attr.Optional = true
				attr.Computed = true
				attr.Description += " Defaults to SQL Server instance's default collation."
				return attr
			}(),
		},
	}, nil
}

func (d databaseResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data, _ := getDB(ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	db := sql.CreateDatabase(ctx, d.Db, data.toSettings())
	if utils.HasError(ctx) {
		return
	}

	settings := db.GetSettings(ctx)
	if utils.HasError(ctx) {
		return
	}

	data = data.withSettings(settings)
	data.Id = types.String{Value: fmt.Sprint(db.GetId(ctx))}
	utils.SetData(ctx, &response.State, data)
}

func (d databaseResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data, dbId := getDB(ctx, request.State)
	if utils.HasError(ctx) {
		return
	}

	db := sql.GetDatabase(ctx, d.Db, dbId)
	if utils.HasError(ctx) {
		return
	}

	dbExists := db.Exists(ctx)
	if utils.HasError(ctx) {
		return
	}
	if !dbExists {
		response.State.RemoveResource(ctx)
		return
	}

	dbSettings := db.GetSettings(ctx)
	if utils.HasError(ctx) {
		return
	}
	utils.SetData(ctx, &response.State, data.withSettings(dbSettings))
}

func (d databaseResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	state, dbId := getDB(ctx, request.State)
	plan := utils.GetData[databaseResourceData](ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	db := sql.GetDatabase(ctx, d.Db, dbId)
	if utils.HasError(ctx) {
		return
	}

	if state.Name.Value != plan.Name.Value {
		db.Rename(ctx, plan.Name.Value)
		if utils.HasError(ctx) {
			return
		}
	}

	if state.Collation.Value != plan.Collation.Value {
		db.SetCollation(ctx, plan.Collation.Value)
		if utils.HasError(ctx) {
			return
		}
	}

	settings := db.GetSettings(ctx)
	if utils.HasError(ctx) {
		return
	}
	utils.SetData(ctx, &response.State, plan.withSettings(settings))
}

func (d databaseResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	_, dbId := getDB(ctx, request.State)
	if utils.HasError(ctx) {
		return
	}

	db := sql.GetDatabase(ctx, d.Db, dbId)
	if utils.HasError(ctx) {
		return
	}

	db.Drop(ctx)
	if utils.HasError(ctx) {
		return
	}

	response.State.RemoveResource(ctx)
}

func (d databaseResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
