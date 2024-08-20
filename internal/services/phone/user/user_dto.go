package user

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
