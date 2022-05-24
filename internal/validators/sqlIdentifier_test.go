package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestSqlIdentifierValidate(t *testing.T) {
	const validationErrSummary = "Invalid SQL identifier"

	testCases := map[string]validatorTestCase{
		"Wrong type": {
			val:             types.Int64{Value: 2},
			expectedSummary: "Value Conversion Error",
		},
		"Unknown": {
			val: types.String{Unknown: true},
		},
		"Null": {
			val: types.String{Null: true},
		},
		"Valid": {
			val: types.String{Value: "_idenTif@$#_er"},
		},
		"startingWithDigit": {
			val:             types.String{Value: "2ndIdentifier"},
			expectedSummary: validationErrSummary,
		},
		"withSpace": {
			val:             types.String{Value: "has space"},
			expectedSummary: validationErrSummary,
		},
		"forbiddenChar": {
			val:             types.String{Value: "has&inName"},
			expectedSummary: validationErrSummary,
		},
	}

	validatorTests(testCases, sqlIdentifierValidator{}, t)
}
