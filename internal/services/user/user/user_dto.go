package user

import (
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listQueryDto struct {
	status        types.String
	roleID        types.String
	includeFields types.String
	license       types.String
}

type listDto struct {
	users []listDtoUser
}

type listDtoUser struct {
	userID            types.String
	customAttributes  []listDtoUserCustomAttribute
	dept              types.String
	displayName       types.String
	email             types.String
	employeeUniqueID  types.String
	firstName         types.String
	groupIDs          []types.String
	hostKey           types.String
	imGroupIDs        []types.String
	lastClientVersion types.String
	lastLoginTime     timetypes.RFC3339
	lastName          types.String
	planUnitedType    types.String
	pmi               types.Int64
	roleID            types.String
	status            types.String
	timezone          types.String
	userType          types.Int32
	userCreatedAt     timetypes.RFC3339
	verified          types.Int32
}

type listDtoUserCustomAttribute struct {
	key   types.String
	name  types.String
	value types.String
}
