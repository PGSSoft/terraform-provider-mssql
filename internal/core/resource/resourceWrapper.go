package resource

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &resourceWrapper[any]{}
	_ resource.ResourceWithImportState    = &resourceWrapper[any]{}
	_ resource.ResourceWithValidateConfig = &resourceWrapper[any]{}
)

func NewResource[T any](r Resource[T]) func() resource.ResourceWithConfigure {
	return func() resource.ResourceWithConfigure {
		return &resourceWrapper[T]{r: r}
	}
}

type resourceWrapper[T any] struct {
	r   Resource[T]
	ctx core.ResourceContext
}

func (r *resourceWrapper[T]) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = fmt.Sprintf("%s_%s", request.ProviderTypeName, r.r.GetName())
}

func (r *resourceWrapper[T]) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	resourceCtx, ok := request.ProviderData.(core.ResourceContext)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected ResourceContext, got: %T. Please report this issue to the provider developers.", request.ProviderData))
		return
	}

	r.ctx = resourceCtx
}

func (r *resourceWrapper[T]) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	return r.r.GetSchema(utils.WithDiagnostics(ctx, &diags)), diags
}

func (r *resourceWrapper[T]) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	req := CreateRequest[T]{}
	req.monad = utils.StopOnError(ctx).
		Then(func() { req.Conn = r.ctx.ConnFactory(ctx) }).
		Then(func() {
			obj := utils.GetData[types.Object](ctx, request.Plan)
			diags := obj.As(ctx, &req.Plan, types.ObjectAsOptions{
				UnhandledUnknownAsEmpty: true,
				UnhandledNullAsEmpty:    true,
			})
			utils.AppendDiagnostics(ctx, diags...)
		})

	resp := CreateResponse[T]{}
	r.r.Create(ctx, req, &resp)

	req.monad.Then(func() { utils.SetData(ctx, &response.State, resp.State) })
}

func (r *resourceWrapper[T]) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	req := ReadRequest[T]{}
	req.monad = utils.StopOnError(ctx).
		Then(func() { req.Conn = r.ctx.ConnFactory(ctx) }).
		Then(func() { req.State = utils.GetData[T](ctx, request.State) })

	resp := ReadResponse[T]{}
	r.r.Read(ctx, req, &resp)

	req.monad.Then(func() {
		if resp.exists {
			utils.SetData(ctx, &response.State, resp.state)
		} else {
			response.State.RemoveResource(ctx)
		}
	})
}

func (r *resourceWrapper[T]) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	req := UpdateRequest[T]{}
	req.monad = utils.StopOnError(ctx).
		Then(func() { req.Conn = r.ctx.ConnFactory(ctx) }).
		Then(func() { req.Plan = utils.GetData[T](ctx, request.Plan) }).
		Then(func() { req.State = utils.GetData[T](ctx, request.State) })

	resp := UpdateResponse[T]{}
	r.r.Update(ctx, req, &resp)

	req.monad.Then(func() { utils.SetData(ctx, &response.State, resp.State) })
}

func (r *resourceWrapper[T]) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	req := DeleteRequest[T]{}
	req.monad = utils.StopOnError(ctx).
		Then(func() { req.Conn = r.ctx.ConnFactory(ctx) }).
		Then(func() { req.State = utils.GetData[T](ctx, request.State) })

	resp := DeleteResponse[T]{}
	r.r.Delete(ctx, req, &resp)

	req.monad.Then(func() { response.State.RemoveResource(ctx) })
}

func (r *resourceWrapper[T]) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *resourceWrapper[T]) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	res, ok := r.r.(ResourceWithValidation[T])
	if !ok {
		return
	}

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	req := ValidateRequest[T]{}
	req.monad = utils.StopOnError(ctx).Then(func() { req.Config = utils.GetData[T](ctx, request.Config) })
	resp := ValidateResponse[T]{}
	res.Validate(ctx, req, &resp)
}
