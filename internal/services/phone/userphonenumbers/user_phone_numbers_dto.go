package userphonenumbers

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	phoneNumbers []*readDtoPhoneNumber
}

type readDtoPhoneNumber struct {
	id     types.String
	number types.String
}

type assignDto struct {
	userID         types.String
	phoneNumberIDs []types.String
	phoneNumbers   []types.String
}

type unassignDto struct {
	userID         types.String
	phoneNumberIDs []types.String
}
