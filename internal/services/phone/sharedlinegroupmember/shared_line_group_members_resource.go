package sharedlinegroupmember

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var (
	_ resource.Resource                = &tfResource{}
	_ resource.ResourceWithConfigure   = &tfResource{}
	_ resource.ResourceWithImportState = &tfResource{}
)

func NewPhoneSharedLineGroupMembersResource() resource.Resource {
	return &tfResource{}
}

type tfResource struct {
	crud *crud
}

func (r *tfResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.crud = newCrud(data.PhoneClient)
}

func (r *tfResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_shared_line_group_members"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `A [shared line group](https://support.zoom.us/hc/en-us/articles/360038850792) allows Zoom Phone admins to share a phone number and extension with a group of phone users or common areas. This gives members of the shared line group access to the group's direct phone number and voicemail. Note that a member can only be added to one shared line group.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:list_users:admin`",
			"`phone:read:list_shared_line_group_members:admin`",
			"`phone:write:shared_line_group_member:admin`",
			"`phone:delete:shared_line_group_member:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"shared_line_group_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier of the Call Queue.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"common_areas": schema.SetNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Common Area.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Common Area ID: Unique identifier of the common area.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the common area.",
						},
						"extension_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The extension ID of the common area.",
						},
					},
				},
			},
			"users": schema.SetNestedAttribute{
				Optional:            true,
				MarkdownDescription: "User.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "User ID: Unique identifier of the user. `id` or `email` must be specified.",
						},
						"email": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Email address of the user. `id` or `email` must be specified.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the user.",
						},
						"extension_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The extension ID of the user.",
						},
					},
				},
			},
		},
	}
}

type resourceModel struct {
	SharedLineGroupID types.String               `tfsdk:"shared_line_group_id"`
	CommonAreas       []*resourceModelCommonArea `tfsdk:"common_areas"`
	Users             []*resourceModelUser       `tfsdk:"users"`
}

type resourceModelCommonArea struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ExtensionID types.String `tfsdk:"extension_id"`
}

type resourceModelUser struct {
	ID          types.String `tfsdk:"id"`
	Email       types.String `tfsdk:"email"`
	Name        types.String `tfsdk:"name"`
	ExtensionID types.String `tfsdk:"extension_id"`
}

func (r *tfResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.read(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone shared line group members", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, plan resourceModel) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, plan.SharedLineGroupID)
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	userExtensionIDs := lo.Map(dto.users, func(member *readDtoUser, _index int) types.String {
		return member.extensionID
	})
	userDatas, err := r.crud.readUsersByExtensionIDs(ctx, userExtensionIDs)
	if err != nil {
		return nil, err
	}

	commonAreas := lo.Ternary(plan.CommonAreas != nil, make([]*resourceModelCommonArea, 0), nil)
	users := lo.Ternary(plan.Users != nil, make([]*resourceModelUser, 0), nil)
	for _, commonArea := range dto.commonAreas {
		commonAreas = append(commonAreas, &resourceModelCommonArea{
			ID:          commonArea.id,
			Name:        commonArea.name,
			ExtensionID: commonArea.extensionID,
		})
	}
	for _, user := range dto.users {
		foundUser, ok := lo.Find(userDatas.users, func(item *readUsersDtoUser) bool {
			return item.extensionID.ValueString() == user.extensionID.ValueString()
		})
		if !ok {
			return nil, fmt.Errorf("user not found: %s", user.extensionID.ValueString())
		}
		users = append(users, &resourceModelUser{
			ID:          user.id,
			Email:       foundUser.email,
			Name:        user.name,
			ExtensionID: user.extensionID,
		})
	}

	return &resourceModel{
		SharedLineGroupID: plan.SharedLineGroupID,
		CommonAreas:       commonAreas,
		Users:             users,
	}, nil
}

func (r *tfResource) sync(ctx context.Context, plan resourceModel) error {
	asis, err := r.read(ctx, plan)
	if err != nil {
		return err
	}

	// 0. plan validation (it might be better to move into validator)
	for _, planUser := range plan.Users {
		if planUser.ID.ValueString() == "" && planUser.Email.ValueString() == "" {
			return fmt.Errorf("either `id` or `email` must be specified on user")
		}
	}

	// 1. unassign members = asis - plan
	var unassignMemberIDs []types.String
	for _, asisUser := range asis.Users {
		planExisted := lo.ContainsBy(plan.Users, func(planItem *resourceModelUser) bool {
			// user parameter allow either id or email
			return planItem.ID == asisUser.ID || planItem.Email == asisUser.Email
		})
		if !planExisted {
			// member id is same with user.id or commonArea.id
			unassignMemberIDs = append(unassignMemberIDs, asisUser.ID)
		}
	}
	for _, asisCommonArea := range asis.CommonAreas {
		planExisted := lo.ContainsBy(plan.CommonAreas, func(planItem *resourceModelCommonArea) bool {
			return planItem.ID == asisCommonArea.ID
		})
		if !planExisted {
			// member id is same with user.id or commonArea.id
			unassignMemberIDs = append(unassignMemberIDs, asisCommonArea.ID)
		}
	}
	if err = r.crud.unassign(ctx, &unassignDto{
		sharedLineGroupID: plan.SharedLineGroupID,
		memberIDs:         unassignMemberIDs,
	}); err != nil {
		return err
	}

	// 2. assign members = plan - asis
	var assignUsers []*assignDtoUser
	for _, planUser := range plan.Users {
		asisExisted := lo.ContainsBy(asis.Users, func(asisItem *resourceModelUser) bool {
			// user parameter allow either id or email
			return asisItem.ID == planUser.ID || asisItem.Email == planUser.Email
		})
		if !asisExisted {
			assignUsers = append(assignUsers, &assignDtoUser{
				id:    planUser.ID, // member id is same with user.id or commonArea.id
				email: planUser.Email,
			})
		}
	}
	var assignCommonAreaIDs []types.String
	for _, planCommonArea := range plan.CommonAreas {
		asisExisted := lo.ContainsBy(asis.CommonAreas, func(asisItem *resourceModelCommonArea) bool {
			return asisItem.ID == planCommonArea.ID
		})
		if !asisExisted {
			// member id is same with user.id or commonArea.id
			assignCommonAreaIDs = append(assignCommonAreaIDs, planCommonArea.ID)
		}
	}
	if err = r.crud.assign(ctx, &assignDto{
		sharedLineGroupID: plan.SharedLineGroupID,
		commonAreaIDs:     assignCommonAreaIDs,
		users:             assignUsers,
	}); err != nil {
		return err
	}
	return nil
}

func (r *tfResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.sync(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone shared line group members",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone shared line group members on reading", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error updating phone shared line group members on get plan",
			"Error updating phone shared line group members",
		)
		return
	}

	if err := r.sync(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone shared line group members",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone shared line group members", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.crud.unassignAll(ctx, state.SharedLineGroupID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone shared line group members",
			fmt.Sprintf(
				"Could not delete phone shared line group members %s, unexpected error: %s",
				state.SharedLineGroupID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone shared line group members", map[string]interface{}{
		"shared_line_group_id": state.SharedLineGroupID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("shared_line_group_id"), req, resp)
}
