package sharedlinegroupgroupphonenumbers

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	primaryNumber types.String
	phoneNumbers  []*readDtoPhoneNumber
}

type readDtoPhoneNumber struct {
	id     types.String
	number types.String
}

type assignDto struct {
	sharedLineGroupID types.String
	phoneNumberIDs    []types.String
	phoneNumbers      []types.String
}

type unassignDto struct {
	sharedLineGroupID types.String
	phoneNumberIDs    []types.String
}
