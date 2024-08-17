package autoreceptionist

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	autoReceptionistID  types.String
	costCenter          types.String
	department          types.String
	extensionID         types.String
	extensionNumber     types.Int64
	name                types.String
	timezone            types.String
	audioPromptLanguage types.String
}

type createDto struct {
	name types.String
}

type createdDto struct {
	autoReceptionistID types.String
	name               types.String
	extensionNumber    types.Int64
}

type updateDto struct {
	autoReceptionistID  types.String
	costCenter          types.String
	department          types.String
	extensionNumber     types.Int64
	name                types.String
	timezone            types.String
	audioPromptLanguage types.String
}
