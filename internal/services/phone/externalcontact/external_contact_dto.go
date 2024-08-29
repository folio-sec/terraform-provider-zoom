package externalcontact

import "github.com/hashicorp/terraform-plugin-framework/types"

type readDto struct {
	description       types.String
	email             types.String
	extensionNumber   types.String
	externalContactID types.String
	id                types.String
	name              types.String
	phoneNumbers      []types.String
	routingPath       types.String
	autoCallRecorded  types.Bool
}

type createDto struct {
	description       types.String
	email             types.String
	extensionNumber   types.String
	externalContactID types.String
	id                types.String
	name              types.String
	phoneNumbers      []types.String
	routingPath       types.String
	autoCallRecorded  types.Bool
}

type createdDto struct {
	externalContactID types.String
}

type updateDto struct {
	externalContactID types.String
	description       types.String
	email             types.String
	extensionNumber   types.String
	id                types.String
	name              types.String
	phoneNumbers      []types.String
	routingPath       types.String
	autoCallRecorded  types.Bool
}
