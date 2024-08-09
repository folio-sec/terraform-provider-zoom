package autoreceptionist

import "github.com/hashicorp/terraform-plugin-framework/types"

type PhoneAutoReceptionistCreateDto struct {
	Name types.String
}

type PhoneAutoReceptionistCreatedDto struct {
	ID              types.String
	Name            types.String
	ExtensionNumber types.Int64
}

type PhoneAutoReceptionistUpdateDto struct {
	AutoReceptionistID  types.String
	CostCenter          types.String
	Department          types.String
	ExtensionNumber     types.Int64
	Name                types.String
	Timezone            types.String
	AudioPromptLanguage types.String
}
