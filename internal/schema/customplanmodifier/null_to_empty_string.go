package customplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nullToEmptyStringPlanModifier struct{}

var _ planmodifier.String = (*nullToEmptyStringPlanModifier)(nil)

func EmptyIfNull() planmodifier.String {
	return &nullToEmptyStringPlanModifier{}
}

func (m *nullToEmptyStringPlanModifier) Description(_ context.Context) string {
	return "Ensures a null string does not trigger diffs on planned values with different cases."
}

func (m *nullToEmptyStringPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m *nullToEmptyStringPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	resp.PlanValue = types.StringValue("")
}
