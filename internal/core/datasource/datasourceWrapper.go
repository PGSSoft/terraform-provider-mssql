package datasource

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func NewDataSource[T any](d DataSource[T]) func() datasource.DataSourceWithConfigure {
	return func() datasource.DataSourceWithConfigure {
		return &dataSourceWrapper[T]{d: d}
	}
}

type dataSourceWrapper[T any] struct {
	d    DataSource[T]
	conn sql.Connection
}

func (d *dataSourceWrapper[T]) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, d.d.GetName())
}

func (d *dataSourceWrapper[T]) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	db, ok := req.ProviderData.(sql.Connection)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected sql.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	d.conn = db
}

func (d *dataSourceWrapper[T]) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	return d.d.GetSchema(utils.WithDiagnostics(ctx, &diags)), diags
}

func (d *dataSourceWrapper[T]) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = utils.WithDiagnostics(ctx, &resp.Diagnostics)

	request := ReadRequest[T]{
		Conn: d.conn,
	}
	request.monad = utils.StopOnError(ctx).Then(func() { request.Config = utils.GetData[T](ctx, req.Config) })

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
