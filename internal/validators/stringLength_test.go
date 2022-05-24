package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestStringLengthValidate(t *testing.T) {
	const validationErrSummary = "Invalid String Length"

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
			val: types.String{Value: "xxxxx"},
		},
		"TooShort": {
			val:             types.String{Value: "xx"},
			expectedSummary: validationErrSummary,
		},
		"TooLong": {
			val:             types.String{Value: "xxxxx xxxxx"},
			expectedSummary: validationErrSummary,
		},
	}

	validatorTests(testCases, stringLengthValidator{Min: 5, Max: 10}, t)
}
