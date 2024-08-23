package user

import (
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceModel struct {
	Status        types.String          `tfsdk:"status"`
	RoleID        types.String          `tfsdk:"role_id"`
	IncludeFields types.String          `tfsdk:"include_fields"`
	License       types.String          `tfsdk:"license"`
	Users         []dataSourceModelUser `tfsdk:"users"`
}

type dataSourceModelUser struct {
	ID                types.String                          `tfsdk:"id"`
	CustomAttributes  []dataSourceModelUserCustomAttributes `tfsdk:"custom_attributes"`
	Dept              types.String                          `tfsdk:"dept"`
	DisplayName       types.String                          `tfsdk:"display_name"`
	Email             types.String                          `tfsdk:"email"`
	EmployeeUniqueID  types.String                          `tfsdk:"employee_unique_id"`
	FirstName         types.String                          `tfsdk:"first_name"`
	GroupIDs          types.Set                             `tfsdk:"group_ids"`
	HostKey           types.String                          `tfsdk:"host_key"`
	ImGroupIDs        types.Set                             `tfsdk:"im_group_ids"`
	LastClientVersion types.String                          `tfsdk:"last_client_version"`
	LastLoginTime     timetypes.RFC3339                     `tfsdk:"last_login_time"`
	LastName          types.String                          `tfsdk:"last_name"`
	PlanUnitedType    types.String                          `tfsdk:"plan_united_type"`
	Pmi               types.Int64                           `tfsdk:"pmi"`
	RoleID            types.String                          `tfsdk:"role_id"`
	Status            types.String                          `tfsdk:"status"`
	Timezone          types.String                          `tfsdk:"timezone"`
	Type              types.Int32                           `tfsdk:"type"`
	UserCreatedAt     timetypes.RFC3339                     `tfsdk:"user_created_at"`
	Verified          types.Int32                           `tfsdk:"verified"`
}

type dataSourceModelUserCustomAttributes struct {
	Key   types.String `tfsdk:"key"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
