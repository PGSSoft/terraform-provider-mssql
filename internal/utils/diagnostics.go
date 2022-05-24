package utils

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const mssqlProviderUtilsDiagnosticsKey = "terraform.mssql.utils.diagnostics"

func GetDiagnostics(ctx context.Context) *diag.Diagnostics {
	return ctx.Value(mssqlProviderUtilsDiagnosticsKey).(*diag.Diagnostics)
}

func WithDiagnostics(ctx context.Context, diagnostics *diag.Diagnostics) context.Context {
	return context.WithValue(ctx, mssqlProviderUtilsDiagnosticsKey, diagnostics)
}

func HasError(ctx context.Context) bool {
	return GetDiagnostics(ctx).HasError()
}

func AddError(ctx context.Context, summary string, err error) {
	GetDiagnostics(ctx).AddError(summary, err.Error())
}

func AppendDiagnostics(ctx context.Context, diagnostics ...diag.Diagnostic) {
	GetDiagnostics(ctx).Append(diagnostics...)
}
