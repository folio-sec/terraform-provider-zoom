package callqueuemember

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	callQueueMembers []*readDtoCallQueueMember
}

type readDtoCallQueueMember struct {
	id          types.String // same with user.id or commonArea.id
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
	callQueueID   types.String
	commonAreaIDs []types.String
	users         []*assignDtoUser
}

type assignDtoUser struct {
	id    types.String
	email types.String
}

type unassignDto struct {
	callQueueID types.String
	memberIDs   []types.String
}
