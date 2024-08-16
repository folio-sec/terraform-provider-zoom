package callqueuemembers

import (
	"context"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var (
	_ resource.Resource              = &tfResource{}
	_ resource.ResourceWithConfigure = &tfResource{}
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
	r.crud = newCrud(data.PhoneMasterClient)
}

func (r *tfResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_call_queue_members"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Call queues allow you to route incoming calls to a group of users. For instance, you can use call queue memberss to route calls to various departments in your organization such as sales, engineering, billing, customer service etc.",
		Attributes: map[string]schema.Attribute{
			"call_queue_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier of the Call Queue.",
			},
			"common_areas": schema.SetNestedAttribute{
				Required:            true,
				MarkdownDescription: "Common Area.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The member ID.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the common area.",
							PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"extension_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The extension ID of the common area.",
							PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"receive_call": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the user can receive calls. It displays if the level is user.",
							PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
						},
					},
				},
			},
			"users": schema.SetNestedAttribute{
				Required:            true,
				MarkdownDescription: "User.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "User ID: Unique identifier of the user.",
						},
						"email": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Email address of the user.",
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

func (r *tfResource) read(ctx context.Context, asis resourceModel) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, asis.CallQueueID)
	if err != nil {
		return nil, fmt.Errorf("error read: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	// check only the members that are in the state
	// sometimes phone_call_queue_members resource is managed by a different state
	// at that time this resource should manage only the members that are in the state
	targetMembers := lo.Filter(dto.callQueueMembers, func(member *readDtoCallQueueMembers, _index int) bool {
		isExistedUser := lo.ContainsBy(asis.Users, func(t *resourceModelUser) bool {
			return t.ID == member.id
		})
		isExistedCommonArea := lo.ContainsBy(asis.CommonAreas, func(t *resourceModelCommonArea) bool {
			return t.ID == member.id
		})
		return isExistedUser || isExistedCommonArea
	})

	commonAreas := make([]*resourceModelCommonArea, 0)
	users := make([]*resourceModelUser, 0)
	for _, member := range targetMembers {
		switch member.level.ValueString() {
		case "user":
			foundUser, ok := lo.Find(asis.Users, func(item *resourceModelUser) bool {
				return item.ID.ValueString() == member.id.ValueString()
			})
			if !ok {
				return nil, fmt.Errorf("user not found: %s", member.id)
			}
			tflog.Info(ctx, "read member", map[string]interface{}{
				"ID":    member.id.ValueString(),
				"Email": foundUser.Email.ValueString(),
				"Name":  member.name.ValueString(),
			})
			users = append(users, &resourceModelUser{
				ID:          member.id,
				Email:       foundUser.Email,
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
		CallQueueID: asis.CallQueueID,
		CommonAreas: commonAreas,
		Users:       users,
	}, nil
}

func (r *tfResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	asis, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call queue members on read",
			err.Error(),
		)
		return
	}

	assignCommonAreaIDs := lo.Map(lo.Filter(plan.CommonAreas, func(t *resourceModelCommonArea, _index int) bool {
		return !lo.ContainsBy(asis.CommonAreas, func(item *resourceModelCommonArea) bool {
			return item.ID == t.ID
		})
	}), func(item *resourceModelCommonArea, index int) types.String {
		return item.ID
	})
	assignUsers := lo.Map(lo.Filter(plan.Users, func(t *resourceModelUser, _index int) bool {
		return !lo.ContainsBy(asis.Users, func(item *resourceModelUser) bool {
			return item.ID == t.ID
		})
	}), func(item *resourceModelUser, index int) *assignDtoUser {
		return &assignDtoUser{
			id:    item.ID,
			email: item.Email,
		}
	})

	if err = r.crud.assign(ctx, &assignDto{
		callQueueID:   plan.CallQueueID,
		commonAreaIDs: assignCommonAreaIDs,
		users:         assignUsers,
	}); err != nil {
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
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error updating phone call queue members on get state",
			"Error updating phone call queue members",
		)
		return
	}
	var plan resourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error updating phone call queue members on get plan",
			"Error updating phone call queue members",
		)
		return
	}
	asis, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call queue members on read",
			err.Error(),
		)
		return
	}

	// 1. unassign members = (state - plan) ∧ asis
	unassignUsers := lo.Filter(state.Users, func(stateItem *resourceModelUser, index int) bool {
		isUnassignPlan := !lo.ContainsBy(plan.Users, func(planItem *resourceModelUser) bool {
			return planItem.ID == stateItem.ID
		})
		isAssigned := lo.ContainsBy(asis.Users, func(asisItem *resourceModelUser) bool {
			return asisItem.ID == stateItem.ID
		})
		return isUnassignPlan && isAssigned
	})
	unassignCommonAreas := lo.Filter(state.CommonAreas, func(stateItem *resourceModelCommonArea, index int) bool {
		isUnassignPlan := !lo.ContainsBy(plan.CommonAreas, func(planItem *resourceModelCommonArea) bool {
			return planItem.ID == stateItem.ID
		})
		isAssigned := lo.ContainsBy(asis.CommonAreas, func(asisItem *resourceModelCommonArea) bool {
			return asisItem.ID == stateItem.ID
		})
		return isUnassignPlan && isAssigned
	})
	var unassignMemberIDs []types.String
	for _, user := range unassignUsers {
		unassignMemberIDs = append(unassignMemberIDs, user.ID)
	}
	for _, commonArea := range unassignCommonAreas {
		unassignMemberIDs = append(unassignMemberIDs, commonArea.ID)
	}
	if err := r.crud.unassign(ctx, &unassignDto{
		callQueueID: plan.CallQueueID,
		memberIDs:   unassignMemberIDs,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone call queue members on unassign",
			fmt.Sprintf(
				"Could not update phone call queue members %s, unexpected error: %s",
				plan.CallQueueID.ValueString(),
				err,
			),
		)
		return
	}

	// 2. assign new members = plan - asis
	assignUsers := lo.Map(lo.Filter(plan.Users, func(planItem *resourceModelUser, index int) bool {
		return !lo.ContainsBy(asis.Users, func(stateItem *resourceModelUser) bool {
			return planItem.ID == stateItem.ID
		})
	}), func(item *resourceModelUser, index int) *assignDtoUser {
		return &assignDtoUser{
			id:    item.ID,
			email: item.Email,
		}
	})
	assignCommonAreaIDs := lo.Map(lo.Filter(plan.CommonAreas, func(planItem *resourceModelCommonArea, index int) bool {
		return !lo.ContainsBy(asis.CommonAreas, func(stateItem *resourceModelCommonArea) bool {
			return planItem.ID == stateItem.ID
		})
	}), func(item *resourceModelCommonArea, index int) types.String {
		return item.ID
	})
	if err := r.crud.assign(ctx, &assignDto{
		callQueueID:   plan.CallQueueID,
		commonAreaIDs: assignCommonAreaIDs,
		users:         assignUsers,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone call queue members on assign",
			fmt.Sprintf(
				"Could not update phone call queue members %s, unexpected error: %s",
				plan.CallQueueID.ValueString(),
				err,
			),
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
	asis, err := r.read(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call queue members on read",
			err.Error(),
		)
		return
	}

	// unassign members = state ∧ asis
	unassignUsers := lo.Filter(state.Users, func(stateItem *resourceModelUser, index int) bool {
		return lo.ContainsBy(asis.Users, func(asisItem *resourceModelUser) bool {
			return asisItem.ID == stateItem.ID
		})
	})
	unassignCommonAreas := lo.Filter(state.CommonAreas, func(stateItem *resourceModelCommonArea, index int) bool {
		return lo.ContainsBy(asis.CommonAreas, func(asisItem *resourceModelCommonArea) bool {
			return asisItem.ID == stateItem.ID
		})
	})
	var unassignMemberIDs []types.String
	for _, user := range unassignUsers {
		unassignMemberIDs = append(unassignMemberIDs, user.ID)
	}
	for _, commonArea := range unassignCommonAreas {
		unassignMemberIDs = append(unassignMemberIDs, commonArea.ID)
	}
	if err := r.crud.unassign(ctx, &unassignDto{
		callQueueID: state.CallQueueID,
		memberIDs:   unassignMemberIDs,
	}); err != nil {
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
