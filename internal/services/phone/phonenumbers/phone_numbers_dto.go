package phonenumbers

import "github.com/hashicorp/terraform-plugin-framework/types"

type readQueryDto struct {
	typ            types.String
	extensionType  types.String
	numberType     types.String
	pendingNumbers types.Bool
	siteID         types.String
}

type readDto struct {
	phoneNumbers []*readDtoPhoneNumber
}

type readDtoPhoneNumber struct {
	assignee                   *readDtoPhoneNumberAssignee
	capability                 []types.String
	carrier                    *readDtoPhoneNumberCarrier
	displayName                types.String
	emergencyAddress           *readDtoPhoneNumberEmergencyAddress
	emergencyAddressStatus     types.Int32
	emergencyAddressUpdateTime types.String
	id                         types.String
	location                   types.String
	number                     types.String
	numberType                 types.String
	sipGroup                   *readDtoPhoneNumberSipGroup
	site                       *readDtoPhoneNumberSite
	source                     types.String
	status                     types.String
}

type readDtoPhoneNumberAssignee struct {
	extensionNumber types.Int64
	id              types.String
	name            types.String
	typ             types.String
}

type readDtoPhoneNumberCarrier struct {
	code types.Int32
	name types.String
}

type readDtoPhoneNumberEmergencyAddress struct {
	addressLine1 types.String
	addressLine2 types.String
	city         types.String
	country      types.String
	stateCode    types.String
	zip          types.String
}

type readDtoPhoneNumberSipGroup struct {
	displayName types.String
	id          types.String
}

type readDtoPhoneNumberSite struct {
	id   types.String
	name types.String
}
