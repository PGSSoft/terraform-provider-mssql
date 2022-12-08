package validators

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
)

// To ensure validator fully satisfies framework interfaces
var _ validator.String = sqlIdentifierValidator{}

type sqlIdentifierValidator struct{}

func (s sqlIdentifierValidator) Description(context.Context) string {
	return "SQL identifier allows letters, digits, @, $, # or _, start with letter, _, @ or #"
}

func (s sqlIdentifierValidator) MarkdownDescription(context.Context) string {
	return "SQL identifier allows letters, digits, `@`, `$`, `#`, `-` or `_`, start with letter, `_`, `@` or `#`. See [MS SQL docs](https://docs.microsoft.com/en-us/sql/relational-databases/databases/database-identifiers) for details."
}

func (s sqlIdentifierValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if !common.IsAttrSet(request.ConfigValue) {
		return
	}

	if match, _ := regexp.Match("^[a-zA-Z_@#][a-zA-Z\\d@$#_-]*$", []byte(request.ConfigValue.ValueString())); !match {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid SQL identifier",
			fmt.Sprintf("%s, got: %s", s.Description(ctx), request.ConfigValue.ValueString()),
		)
	}
}
