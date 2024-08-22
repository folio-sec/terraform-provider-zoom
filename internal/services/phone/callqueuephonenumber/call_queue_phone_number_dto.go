package callqueuephonenumber

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	phoneNumbers []*readDtoPhoneNumber
}

type readDtoPhoneNumber struct {
	id     types.String
	number types.String
	source types.String
}

type assignDto struct {
	callQueueID    types.String
	phoneNumberIDs []types.String
	phoneNumbers   []types.String
}

type unassignDto struct {
	callQueueID    types.String
	phoneNumberIDs []types.String
}
