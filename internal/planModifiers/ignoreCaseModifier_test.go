package planModifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIgnoreCaseModifier(t *testing.T) {
	cases := map[string]struct {
		request       planmodifier.StringRequest
		expectedValue types.String
	}{
		"empty state": {
			request: planmodifier.StringRequest{
				StateValue: types.StringUnknown(),
				PlanValue:  types.StringValue("plannedValue"),
			},
			expectedValue: types.StringValue("plannedValue"),
		},
		"empty plan": {
			request: planmodifier.StringRequest{
				StateValue: types.StringValue("stateValue"),
				PlanValue:  types.StringNull(),
			},
			expectedValue: types.StringNull(),
		},
		"matching case": {
			request: planmodifier.StringRequest{
				StateValue: types.StringValue("matchingCase"),
				PlanValue:  types.StringValue("matchingCase"),
			},
			expectedValue: types.StringValue("matchingCase"),
		},
		"not matching case": {
			request: planmodifier.StringRequest{
				StateValue: types.StringValue("NotMatchingCase"),
				PlanValue:  types.StringValue("NOTMATCHINGCASE"),
			},
			expectedValue: types.StringValue("NotMatchingCase"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			modifier := IgnoreCase()
			response := planmodifier.StringResponse{PlanValue: tc.request.PlanValue}

			modifier.PlanModifyString(context.Background(), tc.request, &response)

			assert.Equal(t, tc.expectedValue, response.PlanValue)
		})
	}
}
