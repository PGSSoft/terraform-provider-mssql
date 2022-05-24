package utils

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type DataGetter interface {
	Get(context.Context, any) diag.Diagnostics
}

func GetData[T any](ctx context.Context, dg DataGetter) (data T) {
	var d T
	diags := dg.Get(ctx, &d)
	AppendDiagnostics(ctx, diags...)
	return d
}

type DataSetter interface {
	Set(context.Context, any) diag.Diagnostics
}

func SetData(ctx context.Context, ds DataSetter, data any) {
	diags := ds.Set(ctx, data)
	AppendDiagnostics(ctx, diags...)
}
