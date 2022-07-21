package planModifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

func IgnoreCase() tfsdk.AttributePlanModifier {
	return ignoreCaseModifier{}
}

type ignoreCaseModifier struct{}

func (m ignoreCaseModifier) Description(context.Context) string {
	return "When config and state string values only differ in casing, the value from state will be used in plan"
}

func (m ignoreCaseModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ignoreCaseModifier) Modify(ctx context.Context, request tfsdk.ModifyAttributePlanRequest, response *tfsdk.ModifyAttributePlanResponse) {
	isNotSet := func(v attr.Value) bool {
		return v == nil || v.IsNull() || v.IsUnknown()
	}

	if isNotSet(request.AttributeState) || isNotSet(request.AttributePlan) {
		return
	}

	if request.AttributePlan.Type(ctx) != types.StringType || request.AttributeState.Type(ctx) != types.StringType {
		return
	}

	toString := func(v attr.Value) *string {
		var strValue string

		value, err := v.ToTerraformValue(ctx)
		if err == nil {
			err = value.As(&strValue)
		}
		if err != nil {
			response.Diagnostics.AddAttributeError(request.AttributePath, "Failed to convert value to string", err.Error())
			return nil
		}

		return &strValue
	}

	plan := toString(request.AttributePlan)
	state := toString(request.AttributeState)

	if plan == nil || state == nil || strings.ToUpper(*plan) != strings.ToUpper(*state) {
		return
	}

	response.AttributePlan = request.AttributeState
}
