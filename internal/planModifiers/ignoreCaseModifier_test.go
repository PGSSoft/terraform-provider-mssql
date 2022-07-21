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
				AttributeState: types.String{Unknown: true},
				AttributePlan:  types.String{Value: "plannedValue"},
			},
			expectedValue: types.String{Value: "plannedValue"},
		},
		"empty plan": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.String{Value: "stateValue"},
				AttributePlan:  types.String{Null: true},
			},
			expectedValue: types.String{Null: true},
		},
		"non string": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.Int64{Value: 246},
				AttributePlan:  types.Int64{Value: 45763},
			},
			expectedValue: types.Int64{Value: 45763},
		},
		"matching case": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.String{Value: "matchingCase"},
				AttributePlan:  types.String{Value: "matchingCase"},
			},
			expectedValue: types.String{Value: "matchingCase"},
		},
		"not matching case": {
			request: tfsdk.ModifyAttributePlanRequest{
				AttributeState: types.String{Value: "NotMatchingCase"},
				AttributePlan:  types.String{Value: "NOTMATCHINGCASE"},
			},
			expectedValue: types.String{Value: "NotMatchingCase"},
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
