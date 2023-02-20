package datasource

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSourceWithValidateConfig = &dataSourceWrapper[any]{}

func NewDataSource[T any](d DataSource[T]) func() datasource.DataSourceWithConfigure {
	return func() datasource.DataSourceWithConfigure {
		return &dataSourceWrapper[T]{d: d}
	}
}

type dataSourceWrapper[T any] struct {
	d   DataSource[T]
	ctx core.ResourceContext
}

func (d *dataSourceWrapper[T]) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, d.d.GetName())
}

func (d *dataSourceWrapper[T]) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	resourceCtx, ok := req.ProviderData.(core.ResourceContext)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected ResourceContext, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	d.ctx = resourceCtx
}

func (d *dataSourceWrapper[T]) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	resp := SchemaResponse{}
	d.d.Schema(utils.WithDiagnostics(ctx, &response.Diagnostics), SchemaRequest{}, &resp)
	response.Schema = resp.Schema
}

func (d *dataSourceWrapper[T]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = utils.WithDiagnostics(ctx, &resp.Diagnostics)

	request := ReadRequest[T]{}
	request.monad = utils.StopOnError(ctx).
		Then(func() { request.Conn = d.ctx.ConnFactory(ctx) }).
		Then(func() { request.Config = utils.GetData[T](ctx, req.Config) })

	response := ReadResponse[T]{}
	d.d.Read(ctx, request, &response)

	request.monad.Then(func() {
		if response.exists {
			utils.SetData(ctx, &resp.State, response.state)
		} else {
			resp.State.RemoveResource(ctx)
		}
	})
}

func (d *dataSourceWrapper[T]) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	dsWithValidation, ok := d.d.(DataSourceWithValidation[T])
	if !ok {
		return
	}

	ctx = utils.WithDiagnostics(ctx, &resp.Diagnostics)

	request := ValidateRequest[T]{}
	request.monad = utils.StopOnError(ctx).
		Then(func() { request.Config = utils.GetData[T](ctx, req.Config) })

	response := ValidateResponse[T]{}
	dsWithValidation.Validate(ctx, request, &response)
}
