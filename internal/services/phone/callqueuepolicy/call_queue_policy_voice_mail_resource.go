package callqueuepolicy

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/samber/lo"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &tfVoiceMailResource{}
	_ resource.ResourceWithConfigure   = &tfVoiceMailResource{}
	_ resource.ResourceWithImportState = &tfVoiceMailResource{}
)

func NewPhoneCallQueuePolicyVoiceMailResource() resource.Resource {
	return &tfVoiceMailResource{}
}

type tfVoiceMailResource struct {
	crud *crud
}

func (r *tfVoiceMailResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tfVoiceMailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_call_queue_policy_voice_mail"
}

func (r *tfVoiceMailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `The policy sub-setting for a specific call queue according to the voice_mail.

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:call_queue:admin`",
			"`phone:write:call_queue_policy:admin`",
			"`phone:update:call_queue_policy:admin`",
			"`phone:delete:call_queue_policy:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"call_queue_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier of the Call Queue.",
			},
			"access_members": schema.SetNestedAttribute{
				Required:            true,
				MarkdownDescription: "The shared voicemail access member list.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"access_user_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The Zoom user ID or email to share or update the access permissions with.",
						},
						"allow_download": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Specifies whether the member has download permissions. The default is **false**.",
						},
						"allow_delete": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Specifies whether the member has delete permissions. The default is **false**.",
						},
						"allow_sharing": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Specifies whether the member has the permission to share. The default is **false**.",
						},
						"shared_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The number is limited to the minimum value of 10 or the number of allowed access members account setting.",
						},
					},
				},
			},
		},
	}
}

type resourceVoiceMailModel struct {
	CallQueueID   types.String                         `tfsdk:"call_queue_id"`
	AccessMembers []resourceVoiceMailModelAccessMember `tfsdk:"access_members"`
}

type resourceVoiceMailModelAccessMember struct {
	AccessUserId  types.String `tfsdk:"access_user_id"`
	AllowDownload types.Bool   `tfsdk:"allow_download"`
	AllowDelete   types.Bool   `tfsdk:"allow_delete"`
	AllowSharing  types.Bool   `tfsdk:"allow_sharing"`
	SharedId      types.String `tfsdk:"shared_id"`
}

func (r *tfVoiceMailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceVoiceMailModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.read(ctx, state.CallQueueID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone call queue policy voice mail", err.Error())
		return
	}

	tflog.Error(ctx, fmt.Sprintf("read output %v", output))
	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfVoiceMailResource) read(ctx context.Context, callQueueID types.String) (*resourceVoiceMailModel, error) {
	dto, err := r.crud.read(ctx, callQueueID)
	if err != nil {
		return nil, fmt.Errorf("error read: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	return &resourceVoiceMailModel{
		CallQueueID: dto.callQueueID,
		AccessMembers: lo.Map(dto.policyVoiceMailMembers, func(item *readDtoPolicyVoiceMailMember, index int) resourceVoiceMailModelAccessMember {
			return resourceVoiceMailModelAccessMember{
				AccessUserId:  item.accessUserID,
				AllowDownload: item.allowDownload,
				AllowDelete:   item.allowDelete,
				AllowSharing:  item.allowSharing,
				SharedId:      item.sharedID,
			}
		}),
	}, nil
}

func (r *tfVoiceMailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceVoiceMailModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.sync(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call queue policy voice mail",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan.CallQueueID)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone call queue voice mail on reading", err.Error())
		return
	}
	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfVoiceMailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceVoiceMailModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error updating phone call queue policy",
			"Error updating phone call queue policy",
		)
		return
	}

	if err := r.sync(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone call queue policy voice mail",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan.CallQueueID)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone call queue voice mail", err.Error())
		return
	}
	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfVoiceMailResource) sync(ctx context.Context, plan resourceVoiceMailModel) error {
	asis, err := r.read(ctx, plan.CallQueueID)
	if err != nil {
		return fmt.Errorf(
			"could not sync phone call queue policy voice mail %s on read, unexpected error: %v",
			plan.CallQueueID.ValueString(),
			err,
		)
	}

	// remove members
	var removeSharedIDs []types.String
	for _, asisMember := range asis.AccessMembers {
		planExisted := lo.ContainsBy(plan.AccessMembers, func(planItem resourceVoiceMailModelAccessMember) bool {
			return planItem.AccessUserId == asisMember.AccessUserId
		})
		if !planExisted {
			removeSharedIDs = append(removeSharedIDs, asisMember.SharedId)
		}
	}
	if len(removeSharedIDs) > 0 {
		if err = r.crud.remove(ctx, &removeDto{
			callQueueID: plan.CallQueueID,
			policyType:  VoiceMail,
			sharedIDs:   removeSharedIDs,
		}); err != nil {
			return fmt.Errorf(
				"could not sync phone call queue policy voice mail %s on remove, unexpected error: %v",
				plan.CallQueueID.ValueString(),
				err,
			)
		}
	}

	// add members
	addMembers := lo.Filter(plan.AccessMembers, func(planItem resourceVoiceMailModelAccessMember, index int) bool {
		return planItem.SharedId.ValueString() == ""
	})
	if len(addMembers) > 0 {
		if err = r.crud.add(ctx, &addDto{
			callQueueID: plan.CallQueueID,
			policyType:  VoiceMail,
			voicemailAccessMembers: lo.Map(addMembers, func(item resourceVoiceMailModelAccessMember, index int) *addDtoVoicemailAccessMember {
				return &addDtoVoicemailAccessMember{
					accessUserID:  item.AccessUserId,
					allowDownload: item.AllowDownload,
					allowDelete:   item.AllowDelete,
					allowSharing:  item.AllowSharing,
				}
			}),
		}); err != nil {
			return fmt.Errorf(
				"could not sync phone call queue policy voice mail %s on add, unexpected error: %v",
				plan.CallQueueID.ValueString(),
				err,
			)
		}
	}

	// update members
	updateMembers := lo.Filter(plan.AccessMembers, func(planItem resourceVoiceMailModelAccessMember, index int) bool {
		return planItem.SharedId.ValueString() != ""
	})
	if err := r.crud.update(ctx, &updateDto{
		callQueueID: plan.CallQueueID,
		policyType:  VoiceMail,
		voicemailAccessMembers: lo.Map(updateMembers, func(item resourceVoiceMailModelAccessMember, index int) *updateDtoVoicemailAccessMember {
			return &updateDtoVoicemailAccessMember{
				accessUserID:  item.AccessUserId,
				allowDownload: item.AllowDownload,
				allowDelete:   item.AllowDelete,
				allowSharing:  item.AllowSharing,
				sharedID:      item.SharedId,
			}
		}),
	}); err != nil {
		return fmt.Errorf(
			"could not update phone call queue policy voice mail %s on update, unexpected error: %v",
			plan.CallQueueID.ValueString(),
			err,
		)
	}
	return nil
}

func (r *tfVoiceMailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceVoiceMailModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	asis, err := r.read(ctx, state.CallQueueID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone call queue policy voice mail on read",
			fmt.Sprintf(
				"Could not delete phone call queue policy %s, unexpected error: %s",
				state.CallQueueID.ValueString(),
				err,
			),
		)
		return
	}
	if asis == nil || len(asis.AccessMembers) == 0 {
		return
	}

	var removeSharedIDs []types.String
	for _, asisMember := range asis.AccessMembers {
		removeSharedIDs = append(removeSharedIDs, asisMember.SharedId)
	}
	if err := r.crud.remove(ctx, &removeDto{
		callQueueID: state.CallQueueID,
		policyType:  VoiceMail,
		sharedIDs:   removeSharedIDs,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone call queue policy voice mail",
			fmt.Sprintf(
				"Could not delete phone call queue policy %s, unexpected error: %s",
				state.CallQueueID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone call queue policy voice mail", map[string]interface{}{
		"call_queue_id": state.CallQueueID.ValueString(),
	})
}

func (r *tfVoiceMailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("call_queue_id"), req, resp)
}
