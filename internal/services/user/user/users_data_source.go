package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ datasource.DataSource              = &tfDataSource{}
	_ datasource.DataSourceWithConfigure = &tfDataSource{}
)

func NewUsersDataSource() datasource.DataSource {
	return &tfDataSource{}
}

type tfDataSource struct {
	crud *crud
}

func (d *tfDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*shared.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected ProviderData Source Configure Type",
			fmt.Sprintf("Expected *provider.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.crud = newCrud(data.UserClient)
}

func (d *tfDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_users"
}

func (d *tfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownSeparatorForList := "\n  "

	resp.Schema = schema.Schema{
		MarkdownDescription: `Gets basic information for multiple Zoom users.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`user:read:list_users:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"query": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "The query parameters for listing users.",
				Attributes: map[string]schema.Attribute{
					"status": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf("active", "inactive", "pending"),
						},
						MarkdownDescription: "The user's status. Default behavior is `active`. Allowed: `active`, `inactive`, `pending`" + strings.Join([]string{
							"",
							"- `active`: The user exists on the account.",
							"- `inactive`: The user has been deactivated.",
							"- `pending`: The user exists on the account, but has not activated their account. See [Managing users](https://support.zoom.us/hc/en-us/articles/201363183) for details.",
						}, markdownSeparatorForList),
					},
					"role_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The role's unique ID. Use this parameter to filter the response by a specific role. You can use the [List roles API](https://developers.zoom.us/docs/api/rest/reference/account/methods/#operation/roles) to get a role's unique ID value.",
					},
					"include_fields": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf("custom_attributes", "host_key"),
						},
						MarkdownDescription: "This parameter displays one of the following attributes in the API call's response. Allowed: `custom_attributes`, `host_key`" + strings.Join([]string{
							"",
							"- `custom_attributes`: Return the user's custom attributes.",
							"- `host_key`: Return the user's [host key](https://support.zoom.us/hc/en-us/articles/205172555-Using-your-host-key).",
						}, markdownSeparatorForList),
					},
					"license": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf("zoom_workforce_management", "zoom_compliance_management"),
						},
						MarkdownDescription: "The user's license. Filter the response by a specific license. Allowed: `zoom_workforce_management`, `zoom_compliance_management`",
					},
				},
			},
			"users": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "A list of users. Each user object provides the attributes documented below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The user's ID. The API does not return this value for users with the `pending` status.",
						},
						"custom_attributes": schema.SetNestedAttribute{
							Computed:            true,
							MarkdownDescription: "The information about the user's custom attributes. This field is only returned if users are assigned custom attributes and you provided the `custom_attributes` value for the `include_fields` parameter in the API request.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The custom attribute's unique ID.",
									},
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The custom attribute's name.",
									},
									"value": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The custom attribute's value.",
									},
								},
							},
						},
						"dept": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user's department.",
						},
						"display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The user's display name.",
						},
						"email": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The user's email address.",
						},
						"employee_unique_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The employee's unique ID. The this field only returns when SAML single sign-on (SSO) is enabled or the `login_type` value is `101` (SSO).",
						},
						"first_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The user's first name.",
						},
						"group_ids": schema.SetAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "The IDs of groups where the user is a member.",
						},
						"host_key": schema.StringAttribute{
							Computed:            true,
							Sensitive:           true,
							MarkdownDescription: "(Optional) The user's [host key](https://support.zoom.us/hc/en-us/articles/205172555-Using-your-host-key). This field is only returned if users are assigned a host key and you provided the host_key value for the include_fields query parameter in the API request.",
						},
						"im_group_ids": schema.SetAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "The IDs of IM directory groups where the user is a member.",
						},
						"last_client_version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The last client version that user used to log in.",
						},
						"last_login_time": schema.StringAttribute{
							CustomType:          timetypes.RFC3339Type{},
							Computed:            true,
							MarkdownDescription: "(Optional) The user's last login time. This field has a three-day buffer period.",
						},
						"last_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The user's last name.",
						},
						"plan_united_type": schema.StringAttribute{
							Computed: true,
							MarkdownDescription: "(Optional) This field is returned if the user is enrolled in the [Zoom United](https://zoom.us/pricing/zoom-bundles) plan. Allowed: `1`, `2`, `4`, `8`, `16`, `32`, `64`, `128`, `256`, `512`, `1024`, `2048`, `4096`, `8192`, `16384`, `32768`, `65536`, `131072`" + strings.Join([]string{
								"",
								"- `1`:  Zoom United Pro-United with US/CA Unlimited.",
								"- `2`:  Zoom United Pro-United with UK/IR Unlimited.",
								"- `4`:  Zoom United Pro-United with AU/NZ Unlimited.",
								"- `8`:  Zoom United Pro-United with Global Select.",
								"- `16`: Zoom United Pro-United with Zoom Phone Pro.",
								"- `32`: Zoom United Biz-United with US/CA Unlimited.",
								"- `64`: Zoom United Biz-United with UK/IR Unlimited.",
								"- `128`: Zoom United Biz-United with AU/NZ Unlimited.",
								"- `256`: Zoom United Biz-United with Global Select.",
								"- `512`: Zoom United Biz-United with Zoom Phone Pro.",
								"- `1024`: Zoom United Ent-United with US/CA Unlimited.",
								"- `2048`: Zoom United Ent-United with UK/IR Unlimited.",
								"- `4096`: Zoom United Ent-United with AU/NZ Unlimited.",
								"- `8192`: Zoom United Ent-United with Global Select.",
								"- `16384`: Zoom United Ent-United with Zoom Phone Pro.",
								"- `32768`: Zoom United Pro-United with JP Unlimited.",
								"- `65536`: Zoom United Biz-United with JP Unlimited.",
								"- `131072`: Zoom United Ent-United with JP Unlimited.",
							}, markdownSeparatorForList),
						},
						"pmi": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The user's [personal meeting ID (PMI)](https://developers.zoom.us/docs/api/rest/using-zoom-apis/#understanding-personal-meeting-id-pmi).",
						},
						"role_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The unique ID of the user's assigned [role](https://developers.zoom.us/docs/api/rest/reference/account/methods/#operation/roles).",
						},
						"status": schema.StringAttribute{
							Computed: true,
							MarkdownDescription: "The user's status. Allowed: `active`, `inactive`, `pending`" + strings.Join([]string{
								"",
								"- `active`: An active user.",
								"- `inactive`: A deactivated user.",
								"- `pending`: A pending user.",
							}, markdownSeparatorForList),
						},
						"timezone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "(Optional) The user's timezone.",
						},
						"type": schema.Int32Attribute{
							Computed: true,
							MarkdownDescription: "The user's assigned plan type. Allowed: `1`, `2`, `4`, `99`" + strings.Join([]string{
								"",
								"- `1`: Basic.",
								"- `2`: Licensed.",
								"- `4`: Unassigned without Meetings Basic.",
								"- `99`: None (this can only be set with `ssoCreate`).",
							}, markdownSeparatorForList),
						},
						"user_created_at": schema.StringAttribute{
							CustomType:          timetypes.RFC3339Type{},
							Computed:            true,
							MarkdownDescription: "The date and time when this user was created.",
						},
						"verified": schema.Int32Attribute{
							Computed: true,
							MarkdownDescription: "Whether the user's email address for the Zoom account is verified. Allowed: `0`, `1`" + strings.Join([]string{
								"- `0`: The user's email not verified.",
								"- `1`: A verified user email.",
							}, markdownSeparatorForList),
						},
					},
				},
			},
		},
	}
}

func (d *tfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dto, err := d.crud.list(ctx, lo.TernaryF(data.Query == nil, func() listQueryDto {
		return listQueryDto{
			status:        types.StringNull(),
			roleID:        types.StringNull(),
			includeFields: types.StringNull(),
			license:       types.StringNull(),
		}
	}, func() listQueryDto {
		return listQueryDto{
			status:        data.Query.Status,
			roleID:        data.Query.RoleID,
			includeFields: data.Query.IncludeFields,
			license:       data.Query.License,
		}
	}))
	if err != nil {
		resp.Diagnostics.AddError("Error reading users", err.Error())
		return
	}

	data.Users = lo.Map(dto.users, func(user listDtoUser, _ int) dataSourceModelUser {
		return dataSourceModelUser{
			ID: lo.Ternary(lo.IsEmpty(user.userID.ValueString()), types.StringNull(), user.userID),
			CustomAttributes: lo.Map(user.customAttributes, func(customAttribute listDtoUserCustomAttribute, _ int) dataSourceModelUserCustomAttributes {
				return dataSourceModelUserCustomAttributes{
					Key:   customAttribute.key,
					Name:  customAttribute.name,
					Value: customAttribute.value,
				}
			}),
			Dept:             user.dept,
			DisplayName:      user.displayName,
			Email:            user.email,
			EmployeeUniqueID: user.employeeUniqueID,
			FirstName:        user.firstName,
			GroupIDs: types.SetValueMust(
				types.StringType,
				// See also: https://go.dev/doc/faq#convert_slice_of_interface
				lo.Map(user.groupIDs, func(v types.String, _ int) attr.Value { return v }),
			),
			HostKey: user.hostKey,
			ImGroupIDs: types.SetValueMust(
				types.StringType,
				// See also: https://go.dev/doc/faq#convert_slice_of_interface
				lo.Map(user.imGroupIDs, func(v types.String, _ int) attr.Value { return v }),
			),
			LastClientVersion: user.lastClientVersion,
			LastLoginTime:     user.lastLoginTime,
			LastName:          user.lastName,
			PlanUnitedType:    user.planUnitedType,
			Pmi:               lo.Ternary(lo.IsEmpty(user.pmi.ValueInt64()), types.Int64Null(), user.pmi),
			RoleID:            lo.Ternary(user.roleID.ValueString() == "0", types.StringNull(), user.roleID),
			Status:            user.status,
			Timezone:          user.timezone,
			Type:              user.userType,
			UserCreatedAt:     user.userCreatedAt,
			Verified:          user.verified,
		}
	})

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
