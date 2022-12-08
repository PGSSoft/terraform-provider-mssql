package validators

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	// To ensure validator fully satisfies framework interfaces
	_ tfsdk.AttributeValidator = stringLengthValidator{}
	_ validator.String         = stringLengthValidator{}
)

type stringLengthValidator struct {
	Min int
	Max int
}

func (s stringLengthValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	s.validateStringImpl(utils.WithDiagnostics(ctx, &response.Diagnostics), request.Path, request.ConfigValue)
}

func (s stringLengthValidator) Description(context.Context) string {
	return fmt.Sprintf("string length must be between %d and %d", s.Min, s.Max)
}

func (s stringLengthValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("string length must be between `%d` and `%d`", s.Min, s.Max)
}

func (s stringLengthValidator) Validate(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, request.AttributeConfig, &str)
	if response.Diagnostics.Append(diags...); diags.HasError() {
		return
	}

	s.validateStringImpl(utils.WithDiagnostics(ctx, &response.Diagnostics), request.AttributePath, str)
}

func (s stringLengthValidator) validateStringImpl(ctx context.Context, path path.Path, str types.String) {
	if !common.IsAttrSet(str) {
		return
	}

	strLen := len(str.ValueString())

	if strLen < s.Min || strLen > s.Max {
		utils.AddAttributeError(ctx, path, "Invalid String Length", fmt.Sprintf("%s, got: %s (%d).", s.Description(ctx), str.ValueString(), strLen))
	}
}
