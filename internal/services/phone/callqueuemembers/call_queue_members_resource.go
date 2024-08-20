package callqueuemembers

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

func NewPhoneCallQueueMembersResource() resource.Resource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_call_queue_members"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Call queues allow you to route incoming calls to a group of users. For instance, you can use call queue members to route calls to various departments in your organization such as sales, engineering, billing, customer service etc.

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:list_users:admin`",
			"`phone:read:list_call_queue_members:admin`",
			"`phone:write:call_queue_member:admin`",
			"`phone:delete:call_queue_member:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"call_queue_id": schema.StringAttribute{
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
						"receive_call": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the user can receive calls. It displays if the level is user.",
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
						"receive_call": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the user can receive calls. It displays if the level is user.",
						},
					},
				},
			},
		},
	}
}

type resourceModel struct {
	CallQueueID types.String               `tfsdk:"call_queue_id"`
	CommonAreas []*resourceModelCommonArea `tfsdk:"common_areas"`
	Users       []*resourceModelUser       `tfsdk:"users"`
}

type resourceModelCommonArea struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ExtensionID types.String `tfsdk:"extension_id"`
	ReceiveCall types.Bool   `tfsdk:"receive_call"`
}

type resourceModelUser struct {
	ID          types.String `tfsdk:"id"`
	Email       types.String `tfsdk:"email"`
	Name        types.String `tfsdk:"name"`
	ExtensionID types.String `tfsdk:"extension_id"`
	ReceiveCall types.Bool   `tfsdk:"receive_call"`
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
		resp.Diagnostics.AddError("Error reading phone call queue members", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, plan resourceModel) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, plan.CallQueueID)
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	extensionIDs := lo.Map(dto.callQueueMembers, func(member *readDtoCallQueueMember, _index int) types.String {
		return member.extensionID
	})
	userDatas, err := r.crud.readUsersByExtensionIDs(ctx, extensionIDs)
	if err != nil {
		return nil, err
	}

	commonAreas := lo.Ternary(plan.CommonAreas != nil, make([]*resourceModelCommonArea, 0), nil)
	users := lo.Ternary(plan.Users != nil, make([]*resourceModelUser, 0), nil)
	for _, member := range dto.callQueueMembers {
		switch member.level.ValueString() {
		case "user":
			foundUser, ok := lo.Find(userDatas.users, func(item *readUsersDtoUser) bool {
				return item.extensionID.ValueString() == member.extensionID.ValueString()
			})
			if !ok {
				return nil, fmt.Errorf("user not found: %s", member.extensionID.ValueString())
			}
			users = append(users, &resourceModelUser{
				ID:          member.id,
				Email:       foundUser.email,
				Name:        member.name,
				ExtensionID: member.extensionID,
				ReceiveCall: member.receiveCall,
			})
			break
		case "commonArea":
			commonAreas = append(commonAreas, &resourceModelCommonArea{
				ID:          member.id,
				Name:        member.name,
				ExtensionID: member.extensionID,
				ReceiveCall: member.receiveCall,
			})
			break
		default:
			return nil, fmt.Errorf("unexpected level: %s", member.level)
		}
	}

	return &resourceModel{
		CallQueueID: plan.CallQueueID,
		CommonAreas: commonAreas,
		Users:       users,
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
		callQueueID: plan.CallQueueID,
		memberIDs:   unassignMemberIDs,
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
		callQueueID:   plan.CallQueueID,
		commonAreaIDs: assignCommonAreaIDs,
		users:         assignUsers,
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
			"Error creating phone call queue members",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone call queue members on reading", err.Error())
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
			"Error updating phone call queue members on get plan",
			"Error updating phone call queue members",
		)
		return
	}

	if err := r.sync(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call queue members",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone call queue members", err.Error())
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

	if err := r.crud.unassignAll(ctx, state.CallQueueID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone call queue members",
			fmt.Sprintf(
				"Could not delete phone call queue members %s, unexpected error: %s",
				state.CallQueueID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone call queue members", map[string]interface{}{
		"call_queue_id": state.CallQueueID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("call_queue_id"), req, resp)
}
