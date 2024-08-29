package user

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listQueryDto struct {
	siteID      types.String
	callingType types.Int32
	status      types.String
	department  types.String
	costCenter  types.String
	keyword     types.String
}

type listDto struct {
	users []listDtoUser
}

type listDtoUser struct {
	userID          types.String
	phoneUserID     types.String
	name            types.String
	email           types.String
	extensionID     types.String
	extensionNumber types.Int64
	status          types.String
	department      types.String
	costCenter      types.String
	site            *listDtoUserSite
	phoneNumbers    []*listDtoUserPhoneNumber
	callingPlans    []*listDtoUserCallingPlan
}

type listDtoUserSite struct {
	id   types.String
	name types.String
}

type listDtoUserPhoneNumber struct {
	id     types.String
	number types.String
}

type listDtoUserCallingPlan struct {
	name               types.String
	typ                types.Int32
	billingAccountID   types.String
	billingAccountName types.String
}

type readDto struct {
	callingPlans       []readDtoCallingPlan
	costCenter         types.String
	department         types.String
	email              types.String
	emergencyAddressID types.String
	extensionID        types.String
	extensionNumber    types.Int64
	zoomUserID         types.String
	phoneNumbers       []readDtoPhoneNumber
	phoneUserID        types.String
	siteID             types.String
	status             types.String
}

type readDtoCallingPlan struct {
	callingPlanType    types.Int32
	billingAccountID   types.String
	billingAccountName types.String
}

type readDtoPhoneNumber struct {
	phoneNumberID types.String
	phoneNumber   types.String
}

type createDto struct {
	zoomUserID types.String
}

type createdDto struct {
}

type updateDto struct {
	zoomUserID         types.String
	emergencyAddressID types.String
	extensionNumber    types.Int64
	siteID             types.String
	templateID         types.String
}
