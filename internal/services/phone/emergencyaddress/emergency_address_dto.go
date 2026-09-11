package emergencyaddress

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type createDto struct {
	addressLine1 types.String
	addressLine2 types.String
	city         types.String
	country      types.String
	isDefault    types.Bool
	siteID       types.String
	state        types.String
	zip          types.String
}

type updateDto struct {
	id           types.String
	addressLine1 types.String
	addressLine2 types.String
	city         types.String
	country      types.String
	isDefault    types.Bool
	siteID       types.String
	state        types.String
	zip          types.String
}

type readDto struct {
	id           types.String
	addressLine1 types.String
	addressLine2 types.String
	city         types.String
	country      types.String
	isDefault    types.Bool
	level        types.Int32
	site         readDtoSite
	owner        readDtoOwner
	stateCode    types.String
	status       types.Int32
	zip          types.String
}

type readDtoSite struct {
	ID   types.String
	Name types.String
}

type readDtoOwner struct {
	ID              types.String
	ExtensionNumber types.Int64
	Name            types.String
}
