package blockedlist

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	blockedListID types.String
	blockType     types.String
	comment       types.String
	matchType     types.String
	phoneNumber   types.String
	status        types.String
}

type createDto struct {
	blockType types.String
	comment   types.String
	// country     types.String read api doesn't support country
	matchType   types.String
	phoneNumber types.String
	status      types.String
}

type createdDto struct {
	blockedListID types.String
}
