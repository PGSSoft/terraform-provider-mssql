package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestSqlIdentifierValidate(t *testing.T) {
	const validationErrSummary = "Invalid SQL identifier"

	testCases := map[string]validatorTestCase{
		"Wrong type": {
			val:             types.Int64Value(2),
			expectedSummary: "Value Conversion Error",
		},
		"Unknown": {
			val: types.StringUnknown(),
		},
		"Null": {
			val: types.StringNull(),
		},
		"Valid": {
			val: types.StringValue("_idenTif@$#_er"),
		},
		"startingWithDigit": {
			val:             types.StringValue("2ndIdentifier"),
			expectedSummary: validationErrSummary,
		},
		"withSpace": {
			val:             types.StringValue("has space"),
			expectedSummary: validationErrSummary,
		},
		"forbiddenChar": {
			val:             types.StringValue("has&inName"),
			expectedSummary: validationErrSummary,
		},
	}

	validatorTests(testCases, sqlIdentifierValidator{}, t)
}
