package planModifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIgnoreCaseModifier(t *testing.T) {
	cases := map[string]struct {
		request       tfsdk.ModifyAttributePlanRequest
		expectedValue attr.Value
	}{
		"empty state": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.StringUnknown(),
				AttributePlan:  types.StringValue("plannedValue"),
			},
			expectedValue: types.StringValue("plannedValue"),
		},
		"empty plan": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.StringValue("stateValue"),
				AttributePlan:  types.StringNull(),
			},
			expectedValue: types.StringNull(),
		},
		"non string": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.Int64Value(246),
				AttributePlan:  types.Int64Value(45763),
			},
			expectedValue: types.Int64Value(45763),
		},
		"matching case": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.StringValue("matchingCase"),
				AttributePlan:  types.StringValue("matchingCase"),
			},
			expectedValue: types.StringValue("matchingCase"),
		},
		"not matching case": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.StringValue("NotMatchingCase"),
				AttributePlan:  types.StringValue("NOTMATCHINGCASE"),
			},
			expectedValue: types.StringValue("NotMatchingCase"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			modifier := IgnoreCase()
			response := tfsdk.ModifyAttributePlanResponse{AttributePlan: tc.request.AttributePlan}

			modifier.Modify(context.Background(), tc.request, &response)

			assert.Equal(t, tc.expectedValue, response.AttributePlan)
		})
	}
}
