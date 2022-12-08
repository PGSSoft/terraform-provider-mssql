package validators

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	// To ensure validator fully satisfies framework interfaces
	_ validator.String = stringLengthValidator{}
)

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

func (s stringLengthValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if !common.IsAttrSet(request.ConfigValue) {
		return
	}

	strLen := len(request.ConfigValue.ValueString())

	if strLen < s.Min || strLen > s.Max {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid String Length",
			fmt.Sprintf("%s, got: %s (%d).", s.Description(ctx), request.ConfigValue.ValueString(), strLen))
	}
}
