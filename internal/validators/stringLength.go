package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure validator fully satisfies framework interfaces
var _ tfsdk.AttributeValidator = stringLengthValidator{}

type stringLengthValidator struct {
	Min int
	Max int
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

	if str.Unknown || str.Null {
		return
	}

	strLen := len(str.Value)

	if strLen < s.Min || strLen > s.Max {
		response.Diagnostics.AddAttributeError(
			request.AttributePath,
			"Invalid String Length",
			fmt.Sprintf("%s, got: %s (%d).", s.Description(ctx), str.Value, strLen),
		)
	}
}
