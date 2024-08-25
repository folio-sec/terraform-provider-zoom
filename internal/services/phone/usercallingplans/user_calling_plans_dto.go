package usercallingplans

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type readDto struct {
	callingPlans []readDtoCallingPlan
}

type readDtoCallingPlan struct {
	callingPlanType  types.Int32
	callingPlanName  types.String
	billingAccountID types.String
}

type createDto struct {
	userID       types.String
	callingPlans []createDtoCallingPlan
}

type createDtoCallingPlan struct {
	callingPlanType  types.Int32
	billingAccountID types.String
}

type createdDto struct{}

type deleteDto struct {
	userID       types.String
	callingPlans []deleteDtoCallingPlan
}

type deleteDtoCallingPlan struct {
	callingPlanType  types.Int32
	billingAccountID types.String
}
