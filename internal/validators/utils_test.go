package validators

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

type validatorTestCase struct {
	val             attr.Value
	expectedSummary string
}

func validatorTests(testCases map[string]validatorTestCase, validator tfsdk.AttributeValidator, t *testing.T) {
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			request := tfsdk.ValidateAttributeRequest{AttributeConfig: tc.val}
			response := tfsdk.ValidateAttributeResponse{}

			validator.Validate(context.Background(), request, &response)

			if tc.expectedSummary == "" {
				assert.Falsef(t, response.Diagnostics.HasError(), "Unexpected validation errrors: %v", response.Diagnostics)
			} else {
				for _, d := range response.Diagnostics {
					if d.Severity() == diag.SeverityError && d.Summary() == tc.expectedSummary {
						return
					}
				}
				t.Errorf("Did not find expected validation error '%s'", tc.expectedSummary)
			}
		})
	}
}
