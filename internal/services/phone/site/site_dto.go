package site

import "github.com/hashicorp/terraform-plugin-framework/types"

type createDto struct {
	autoReceptionistName     types.String
	sourceAutoReceptionistID types.String
	defaultEmergencyAddress  createDtoDefaultEmergencyAddress
	name                     types.String
	shortExtensionLength     types.Int32
	siteCode                 types.Int32
	sipZoneID                types.String
	indiaStateCode           types.String
	indiaCity                types.String
	indiaSdcaNpa             types.String
	indiaEntityName          types.String
}

type createDtoDefaultEmergencyAddress struct {
	addressLine1 types.String
	addressLine2 types.String
	city         types.String
	stateCode    types.String
	zip          types.String
}

type createDtoShortExtensionRange struct {
	rangeFrom types.String
	rangeTo   types.String
}

type createDtoForceOffNet struct {
	enable                                      types.Bool
	allowExtensionOnlyUsersCallUsersOutsideSite types.Bool
}

type createdDto struct {
	id   types.String
	name types.String
}

type updateDto struct {
	id                      types.String
	name                    types.String
	siteCode                types.Int32
	shortExtension          *updateDtoShortExtension
	defaultEmergencyAddress updateDtoDefaultEmergencyAddress
	sipZoneID               types.String
	callerIDName            types.String
}

type updateDtoShortExtension struct {
	length types.Int32
	ranges []updateDtoShortExtensionRange
}

type updateDtoShortExtensionRange struct {
	rangeFrom types.String
	rangeTo   types.String
}

type updateDtoDefaultEmergencyAddress struct {
	addressLine1 types.String
	addressLine2 types.String
	city         types.String
	stateCode    types.String
	zip          types.String
}

type readDto struct {
	id                   types.String
	country              readDtoCountry
	mainAutoReceptionist readDtoMainAutoReceptionist
	name                 types.String
	shortExtensionLength types.Int32
	siteCode             types.Int32
	sipZone              readDtoSipZone
	callerIDName         types.String
	indiaStateCode       types.String
	indiaCity            types.String
	indiaSdcaNpa         types.String
	indiaEntityName      types.String
}

type readDtoCountry struct {
	code types.String
	name types.String
}

type readDtoMainAutoReceptionist struct {
	id              types.String
	name            types.String
	extensionID     types.String
	extensionNumber types.Int64
}

type readDtoSipZone struct {
	id   types.String
	name types.String
}

type readDefaultEmergencyAddressDto struct {
	addressLine1 types.String
	addressLine2 types.String
	city         types.String
	countryCode  types.String
	stateCode    types.String
	zip          types.String
}
