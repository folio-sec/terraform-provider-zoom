package callqueue

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	callQueueID     types.String
	costCenter      types.String
	department      types.String
	extensionID     types.String
	extensionNumber types.Int64
	name            types.String
	site            *readDtoSite
	status          types.String
}

type readDtoSite struct {
	id   types.String
	name types.String
}

type createDto struct {
	name            types.String
	siteID          types.String
	costCenter      types.String
	department      types.String
	extensionNumber types.Int64
	description     types.String
}

type createdDto struct {
	callQueueID types.String
}

type updateDto struct {
	callQueueID     types.String
	siteID          types.String
	costCenter      types.String
	department      types.String
	extensionNumber types.Int64
	name            types.String
	description     types.String
	status          types.String
}
