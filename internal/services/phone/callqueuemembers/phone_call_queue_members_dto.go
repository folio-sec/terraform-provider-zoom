package callqueuemembers

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	callQueueMembers []*readDtoCallQueueMembers
}

type readDtoCallQueueMembers struct {
	id          types.String
	name        types.String
	level       types.String
	receiveCall types.Bool
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
