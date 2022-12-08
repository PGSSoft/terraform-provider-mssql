package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestStringLengthValidate(t *testing.T) {
	const validationErrSummary = "Invalid String Length"

	testCases := map[string]validatorTestCase{
		"Unknown": {
			val: types.StringUnknown(),
		},
		"Null": {
			val: types.StringNull(),
		},
		"Valid": {
			val: types.StringValue("xxxxx"),
		},
		"TooShort": {
			val:             types.StringValue("xx"),
			expectedSummary: validationErrSummary,
		},
		"TooLong": {
			val:             types.StringValue("xxxxx xxxxx"),
			expectedSummary: validationErrSummary,
		},
	}

	validatorTests(testCases, stringLengthValidator{Min: 5, Max: 10}, t)
}
