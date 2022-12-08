package planModifiers

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"strings"
)

func IgnoreCase() planmodifier.String {
	return ignoreCaseModifier{}
}

type ignoreCaseModifier struct{}

func (m ignoreCaseModifier) Description(context.Context) string {
	return "When config and state string values only differ in casing, the value from state will be used in plan"
}

func (m ignoreCaseModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ignoreCaseModifier) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	isNotSet := func(v attr.Value) bool {
		return v == nil || v.IsNull() || v.IsUnknown()
	}

	if isNotSet(request.StateValue) || isNotSet(request.PlanValue) {
		return
	}

	if strings.ToUpper(request.PlanValue.ValueString()) != strings.ToUpper(request.StateValue.ValueString()) {
		return
	}

	response.PlanValue = request.StateValue
}
