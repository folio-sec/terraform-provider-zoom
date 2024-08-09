package autoreceptionist

import "github.com/hashicorp/terraform-plugin-framework/types"

type PhoneAutoReceptionistModel struct {
	AutoReceptionistID  types.String `tfsdk:"auto_receptionist_id"`
	CostCenter          types.String `tfsdk:"cost_center"`
	Department          types.String `tfsdk:"department"`
	ExtensionNumber     types.Int64  `tfsdk:"extension_number"`
	Name                types.String `tfsdk:"name"`
	Timezone            types.String `tfsdk:"timezone"`
	AudioPromptLanguage types.String `tfsdk:"audio_prompt_language"`
}
