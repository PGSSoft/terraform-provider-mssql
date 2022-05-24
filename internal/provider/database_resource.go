package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/tkielar/terraform-provider-mssql/internal/utils"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.ResourceType            = DatabaseResourceType{}
	_ tfsdk.Resource                = databaseResource{}
	_ tfsdk.ResourceWithImportState = databaseResource{}
)

type DatabaseResourceType struct{}

func (d DatabaseResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Manages single database.",
		Attributes: map[string]tfsdk.Attribute{
			"id": func() tfsdk.Attribute {
				attr := databaseAttributes["id"]
				attr.Computed = true
				attr.PlanModifiers = tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				}
				return attr
			}(),
			"name": func() tfsdk.Attribute {
				attr := databaseAttributes["name"]
				attr.Required = true
				return attr
			}(),
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

func (d DatabaseResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) databaseResource {
		return databaseResource{Resource: base}
	})
}

type databaseResource struct {
	Resource
}

func (d databaseResource) Create(ctx context.Context, request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data, _ := getDB(ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	db := d.Db.CreateDatabase(ctx, data.toSettings())
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

func (d databaseResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data, dbId := getDB(ctx, request.State)
	if utils.HasError(ctx) {
		return
	}

	db := d.Db.GetDatabase(ctx, dbId)
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

func (d databaseResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest, response *tfsdk.UpdateResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	state, dbId := getDB(ctx, request.State)
	plan := utils.GetData[databaseResourceData](ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	db := d.Db.GetDatabase(ctx, dbId)
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

func (d databaseResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	_, dbId := getDB(ctx, request.State)
	if utils.HasError(ctx) {
		return
	}

	db := d.Db.GetDatabase(ctx, dbId)
	if utils.HasError(ctx) {
		return
	}

	db.Drop(ctx)
	if utils.HasError(ctx) {
		return
	}

	response.State.RemoveResource(ctx)
}

func (d databaseResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), request, response)
}
