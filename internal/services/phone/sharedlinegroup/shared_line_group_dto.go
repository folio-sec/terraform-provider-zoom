package sharedlinegroup

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	sharedLineGroupID types.String
	displayName       types.String
	extensionID       types.String
	extensionNumber   types.Int64
	primaryNumber     types.String
	status            types.String
	site              *readDtoSite
}

type readDtoSite struct {
	id   types.String
	name types.String
}

type createDto struct {
	displayName     types.String
	siteID          types.String
	extensionNumber types.Int64
	description     types.String
}

type createdDto struct {
	sharedLineGroupID types.String
}

type updateDto struct {
	sharedLineGroupID types.String
	extensionNumber   types.Int64
	displayName       types.String
	status            types.String
}
