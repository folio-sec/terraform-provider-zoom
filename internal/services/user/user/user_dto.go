package user

import "github.com/hashicorp/terraform-plugin-framework/types"

type readQueryDto struct {
	status        types.String
	roleID        types.String
	pageNumber    types.String
	includeFields types.String
	license       types.String
}

type readDto struct {
	user []*readDtoUser
}

type readDtoUser struct {
	id               types.String
	email            types.String
	customAttributes []*readDtoUserCustomAttributes
	dept             types.String
	employeeUniqueID types.String
	firstName        types.String
	lastName         types.String
	groupIds         []types.String
	hostKey          types.String
	imGroupIds       []types.String
	planUnitedType   types.String
	pmi              types.Int64
	roleID           types.String
	status           types.String
	typ              types.Int32
	verified         types.Int32
	displayName      types.String
}

type readDtoUserCustomAttributes struct {
	key   types.String
	name  types.String
	value types.String
}
