package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
)

// To ensure validator fully satisfies framework interfaces
var _ tfsdk.AttributeValidator = sqlIdentifierValidator{}

type sqlIdentifierValidator struct{}

func (s sqlIdentifierValidator) Description(context.Context) string {
	return "SQL identifier allows letters, digits, @, $, # or _, start with letter, _, @ or #"
}

func (s sqlIdentifierValidator) MarkdownDescription(context.Context) string {
	return "SQL identifier allows letters, digits, `@`, `$`, `#` or `_`, start with letter, `_`, `@` or `#`. See [MS SQL docs](https://docs.microsoft.com/en-us/sql/relational-databases/databases/database-identifiers) for details."
}

func (s sqlIdentifierValidator) Validate(ctx context.Context, request tfsdk.ValidateAttributeRequest, response *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, request.AttributeConfig, &str)
	if response.Diagnostics.Append(diags...); response.Diagnostics.HasError() {
		return
	}

	if str.Unknown || str.Null {
		return
	}

	if match, _ := regexp.Match("^[a-zA-Z_@#][a-zA-Z\\d@$#_]*$", []byte(str.Value)); !match {
		response.Diagnostics.AddAttributeError(
			request.AttributePath,
			"Invalid SQL identifier",
			fmt.Sprintf("%s, got: %s", s.Description(ctx), str.Value),
		)
	}
}
