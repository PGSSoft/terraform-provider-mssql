package utils

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	if err != nil {
		GetDiagnostics(ctx).AddError(summary, err.Error())
	}
}

func AddAttributeError(ctx context.Context, path path.Path, summary string, details string) {
	GetDiagnostics(ctx).AddAttributeError(path, summary, details)
}

func AppendDiagnostics(ctx context.Context, diagnostics ...diag.Diagnostic) {
	GetDiagnostics(ctx).Append(diagnostics...)
}
