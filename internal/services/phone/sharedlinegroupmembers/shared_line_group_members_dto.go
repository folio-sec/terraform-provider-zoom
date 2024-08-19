package sharedlinegroupgroupmembers

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	commonAreas []*readDtoCommonArea
	users       []*readDtoUser
}

type readDtoCommonArea struct {
	id          types.String // same with commonArea.id
	name        types.String
	extensionID types.String
}

type readDtoUser struct {
	id          types.String // same with user.id
	name        types.String
	level       types.String
	receiveCall types.Bool
	extensionID types.String
}

type readUsersDto struct {
	users []*readUsersDtoUser
}

type readUsersDtoUser struct {
	email       types.String
	extensionID types.String
}

type assignDto struct {
	sharedLineGroupID types.String
	commonAreaIDs     []types.String
	users             []*assignDtoUser
}

type assignDtoUser struct {
	id    types.String
	email types.String
}

type unassignDto struct {
	sharedLineGroupID types.String
	memberIDs         []types.String
}
