package autoreceptionistivr

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_                 resource.Resource                = &tfResource{}
	_                 resource.ResourceWithConfigure   = &tfResource{}
	_                 resource.ResourceWithImportState = &tfResource{}
	allKeys                                            = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "*", "#"}
	keyActionDisabled                                  = int32(-1)
)

func NewPhoneAutoReceptionistIvrResource() resource.Resource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_auto_receptionist_ivr"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `[interactive voice response (IVR) system](https://support.zoom.us/hc/en-us/articles/360038601971) of the specified auto receptionist.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:auto_receptionist_ivr:admin`",
			"`phone:update:auto_receptionist_ivr:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"auto_receptionist_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the auto receptionist. It can be retrieved from the [List Phone Sites](https://marketplace.zoom.us/docs/api-reference/phone/methods#operation/listPhoneSites) API.",
			},
			"hours_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("business_hours"),
				MarkdownDescription: "The query hours type: business_hours or closed_hours, default business_hours.",
			},
			"holiday_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The auto receptionist holiday hours ID. If both holiday_id and hours_type are passed, holiday_id has a high priority and hours_type is invalid.",
			},
			"audio_prompt": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The audio prompt file ID. If the audio was removed from the user's audio library, it will be marked with a prefix, removed_vWby3OZaQlS1nAdmEAqgwA for example. You can still use this audio ID to get the audio information in [Get an audio item](https://marketplace.zoom.us/docs/api-reference/phone/methods#tag/Audio-Library/operation/GetAudioItem) API.",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The audio prompt file name.",
					},
				},
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"id":   types.StringType,
							"name": types.StringType,
						},
						map[string]attr.Value{
							"id":   types.StringValue(""),
							"name": types.StringValue("Default"),
						},
					),
				),
			},
			"caller_enters_no_action": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "The action if caller enters no action after the prompt played.",
				Attributes: map[string]schema.Attribute{
					"action": schema.Int32Attribute{
						Required: true,
						MarkdownDescription: `The action if caller enters no action after the prompt played.
  - -1 Disconnect the call
  - 2 Forward to the user
  - 4 Forward to the common area
  - 5 Forward to Cisco/Polycom Room
  - 6 Forward to the auto receptionist
  - 7 Forward to the call queue
  - 8 Forward to the shared line group
  - 15 Forward to the Contact Center
`,
					},
					"audio_prompt_repeat": schema.Int32Attribute{
						Required: true,
						MarkdownDescription: `The number of times to repeat the audio prompt.
  - Allowed: 1┃2┃3
`,
					},
					"forward_to": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "",
						Attributes: map[string]schema.Attribute{
							"extension_id": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "The extension ID or contact center setting ID.",
							},
							"display_name": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The display name.",
							},
							"extension_number": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The extension number.",
							},
							"id": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The user, common area, Zoom Room, Cisco/Polycom room, auto receptionist, call queue, or shared line group ID.",
							},
						},
					},
				},
			},
			"key_actions": schema.MapNestedAttribute{
				Required:            true,
				MarkdownDescription: "IVR routing options. The keys are supported: '0'-'9', * and #.",
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
					mapvalidator.KeysAre(stringvalidator.OneOf(allKeys...)),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.Int32Attribute{
							Required: true,
							MarkdownDescription: `The action after clicking the key.
  - For key 0-9
    - 100 Leave voicemail to the current extension
    - 200 Leave voicemail to the user
    - 300 Leave voicemail to the auto receptionist
    - 400 Leave voicemail to the call queue
    - 500 Leave voicemail to the shared line group
    - 2 Forward to the user
    - 3 Forward to Zoom Room
    - 4 Forward to the common area
    - 5 Forward to Cisco/Polycom Room
    - 6 Forward to the auto receptionist
    - 7 Forward to the call queue
    - 8 Forward to the shared line group
    - 9 Forward to external contacts
    - 10 Forward to a phone number
    - 15 Forward to the contact center
    - 16 Forward to the meeting service
    - 17 Forward to the meeting service number
    - -1 Disabled
  - For key * or #
    - 21 Repeat menu greeting
    - 22 Return to the root menu
    - 23 Return to the previous menu
    - -1 Disabled
`,
						},
						"target": schema.SingleNestedAttribute{
							Optional:            true,
							MarkdownDescription: "The route to an extension, phone number, or a contact center.",
							Attributes: map[string]schema.Attribute{
								"extension_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "The extension ID or contact center setting ID.",
								},
								"phone_number": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "The phone number to forward.",
								},
								"display_name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The display name.",
								},
								"extension_number": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The extension number.",
								},
								"id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The user, common area, Zoom Room, Cisco/Polycom room, auto receptionist, call queue, or shared line group ID.",
								},
							},
						},
						"voicemail_greeting": schema.SingleNestedAttribute{
							Optional:            true,
							MarkdownDescription: "The voicemail greeting.",
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "The voicemail greeting file ID. If the audio was removed from the user's audio library, it will be marked with a prefix, `removed_vWby3OZaQlS1nAdmEAqgwA` for example. You can still use this audio ID to get the audio information in [Get an audio item](https://marketplace.zoom.us/docs/api-reference/phone/methods#tag/Audio-Library/operation/GetAudioItem) API.",
								},
								"name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The voicemail greeting file name.",
								},
							},
						},
					},
				},
			},
		},
	}
}

type resourceModel struct {
	AutoReceptionistID   types.String                       `tfsdk:"auto_receptionist_id"`
	HoursType            types.String                       `tfsdk:"hours_type"`
	HolidayID            types.String                       `tfsdk:"holiday_id"`
	AudioPrompt          *resourceModelAudioPrompt          `tfsdk:"audio_prompt"`
	CallerEntersNoAction *resourceModelCallerEntersNoAction `tfsdk:"caller_enters_no_action"`
	KeyActions           map[string]*resourceModelKeyAction `tfsdk:"key_actions"`
}

type resourceModelAudioPrompt struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type resourceModelCallerEntersNoAction struct {
	Action            types.Int32                                 `tfsdk:"action"`
	AudioPromptRepeat types.Int32                                 `tfsdk:"audio_prompt_repeat"`
	ForwardTo         *resourceModelCallerEntersNoActionForwardTo `tfsdk:"forward_to"`
}

type resourceModelCallerEntersNoActionForwardTo struct {
	DisplayName     types.String `tfsdk:"display_name"`
	ExtensionID     types.String `tfsdk:"extension_id"`
	ExtensionNumber types.String `tfsdk:"extension_number"`
	ID              types.String `tfsdk:"id"`
}

type resourceModelKeyAction struct {
	Action            types.Int32                              `tfsdk:"action"`
	Target            *resourceModelKeyActionTarget            `tfsdk:"target"`
	VoicemailGreeting *resourceModelKeyActionVoicemailGreeting `tfsdk:"voicemail_greeting"`
}

type resourceModelKeyActionTarget struct {
	ExtensionID     types.String `tfsdk:"extension_id"`
	PhoneNumber     types.String `tfsdk:"phone_number"`
	DisplayName     types.String `tfsdk:"display_name"`
	ExtensionNumber types.String `tfsdk:"extension_number"`
	ID              types.String `tfsdk:"id"`
}

type resourceModelKeyActionVoicemailGreeting struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
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
		resp.Diagnostics.AddError("Error reading phone auto receptionist ivr", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, model resourceModel) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, model.AutoReceptionistID, model.HoursType, model.HolidayID)
	if err != nil {
		return nil, fmt.Errorf("error read: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	var audioPrompt *resourceModelAudioPrompt
	if dto.audioPrompt != nil && model.AudioPrompt != nil {
		audioPrompt = &resourceModelAudioPrompt{
			ID:   dto.audioPrompt.id,
			Name: dto.audioPrompt.name,
		}
	}
	var callerEntersNoAction *resourceModelCallerEntersNoAction
	if dto.callerEntersNoAction != nil && model.CallerEntersNoAction != nil {
		var forwardTo *resourceModelCallerEntersNoActionForwardTo
		if dto.callerEntersNoAction.forwardTo != nil && model.CallerEntersNoAction.ForwardTo != nil {
			forwardTo = &resourceModelCallerEntersNoActionForwardTo{
				DisplayName:     dto.callerEntersNoAction.forwardTo.displayName,
				ExtensionID:     dto.callerEntersNoAction.forwardTo.extensionID,
				ExtensionNumber: dto.callerEntersNoAction.forwardTo.extensionNumber,
				ID:              dto.callerEntersNoAction.forwardTo.id,
			}
		}
		callerEntersNoAction = &resourceModelCallerEntersNoAction{
			Action:            dto.callerEntersNoAction.action,
			AudioPromptRepeat: dto.callerEntersNoAction.audioPromptRepeat,
			ForwardTo:         forwardTo,
		}
	}
	keyActions := map[string]*resourceModelKeyAction{}
	for _, keyAction := range dto.keyActions {
		var modelKeyAction *resourceModelKeyAction
		for key, action := range model.KeyActions {
			if keyAction.key.ValueString() == key {
				modelKeyAction = action
				break
			}
		}
		if modelKeyAction == nil {
			continue
		}

		var target *resourceModelKeyActionTarget
		if keyAction.target != nil && modelKeyAction.Target != nil {
			target = &resourceModelKeyActionTarget{
				DisplayName:     keyAction.target.displayName,
				ExtensionID:     keyAction.target.extensionID,
				ExtensionNumber: keyAction.target.extensionNumber,
				ID:              keyAction.target.id,
				PhoneNumber:     keyAction.target.phoneNumber,
			}
		}
		var voicemailGreeting *resourceModelKeyActionVoicemailGreeting
		if keyAction.voicemailGreeting != nil && modelKeyAction.VoicemailGreeting != nil {
			voicemailGreeting = &resourceModelKeyActionVoicemailGreeting{
				ID:   keyAction.voicemailGreeting.id,
				Name: keyAction.voicemailGreeting.name,
			}
		}
		keyActions[keyAction.key.ValueString()] = &resourceModelKeyAction{
			Action:            keyAction.action,
			Target:            target,
			VoicemailGreeting: voicemailGreeting,
		}
	}

	// if plan has -1 disable action, zoom api doesn't return disabled key action
	// terraform requires the consistent state, so we need to fill-in the disabled key action
	for key, modelKeyAction := range model.KeyActions {
		if modelKeyAction.Action.ValueInt32() != keyActionDisabled {
			continue
		}
		keyActions[key] = &resourceModelKeyAction{
			Action: types.Int32Value(keyActionDisabled),
		}
	}

	return &resourceModel{
		AutoReceptionistID:   dto.autoReceptionistID,
		HoursType:            dto.hoursType,
		HolidayID:            dto.holidayID,
		AudioPrompt:          audioPrompt,
		CallerEntersNoAction: callerEntersNoAction,
		KeyActions:           keyActions,
	}, nil
}

func (r *tfResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// auto receptionist ivr is created when creating auto receptionist
	// so in create, we only update the auto receptionist ivr
	updateRequest := r.buildUpdateDto(plan)
	tflog.Debug(ctx, "phone auto receptionist ivr create request", map[string]interface{}{
		"plan": plan,
		"req":  updateRequest,
	})
	if err := r.crud.update(ctx, updateRequest); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone auto receptionist ivr",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone auto receptionist ivr on reading", err.Error())
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
			"Error updating phone auto receptionist ivr",
			"Error updating phone auto receptionist ivr",
		)
		return
	}

	updateRequest := r.buildUpdateDto(plan)
	tflog.Debug(ctx, "phone auto receptionist ivr update request", map[string]interface{}{
		"plan": plan,
		"req":  updateRequest,
	})
	if err := r.crud.update(ctx, updateRequest); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone auto receptionist ivr",
			fmt.Sprintf(
				"Could not update phone auto receptionist ivr %s, unexpected error: %s",
				plan.AutoReceptionistID.ValueString(),
				err,
			),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone auto receptionist ivr", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) buildUpdateDto(tobe resourceModel) *updateDto {
	ret := &updateDto{
		autoReceptionistID: tobe.AutoReceptionistID,
		holidayID:          tobe.HolidayID,
		hoursType:          tobe.HoursType,
	}
	if tobe.AudioPrompt != nil {
		ret.audioPromptID = tobe.AudioPrompt.ID
	}
	if tobe.CallerEntersNoAction != nil {
		forwardToExtensionID := types.StringNull()
		if tobe.CallerEntersNoAction.ForwardTo != nil {
			forwardToExtensionID = tobe.CallerEntersNoAction.ForwardTo.ExtensionID
		}
		ret.callerEntersNoAction = &updateDtoCallerEntersNoAction{
			action:               tobe.CallerEntersNoAction.Action,
			auditPromptRepeat:    tobe.CallerEntersNoAction.AudioPromptRepeat,
			forwardToExtensionID: forwardToExtensionID,
		}
	}
	var keyActions []*updateDtoKeyAction
	for _, key := range allKeys { // fill-in all 0-9,#,* key actions to update idempotent
		keyAction := &updateDtoKeyAction{
			key:    types.StringValue(key),
			action: types.Int32Value(-1), // disabled by default = -1
		}
		for tobeKey, tobeAction := range tobe.KeyActions {
			if tobeKey == key {
				var target *updateDtoKeyActionTarget
				if tobeAction.Target != nil {
					target = &updateDtoKeyActionTarget{
						extensionID: tobeAction.Target.ExtensionID,
						phoneNumber: tobeAction.Target.PhoneNumber,
					}
				}
				voiceMailGreetingId := types.StringNull()
				if tobeAction.VoicemailGreeting != nil {
					voiceMailGreetingId = tobeAction.VoicemailGreeting.ID
				}
				keyAction = &updateDtoKeyAction{
					key:                 types.StringValue(key),
					action:              tobeAction.Action,
					target:              target,
					voiceMailGreetingId: voiceMailGreetingId,
				}
				break
			}
		}
		keyActions = append(keyActions, keyAction)
	}
	ret.keyActions = keyActions
	return ret
}

func (r *tfResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.crud.delete(ctx, state.AutoReceptionistID, state.HoursType, state.HolidayID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone auto receptionist ivr",
			fmt.Sprintf(
				"Could not delete phone auto receptionist ivr %s, unexpected error: %s",
				state.AutoReceptionistID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone auto receptionist ivr", map[string]interface{}{
		"auto_receptionist_id": state.AutoReceptionistID.ValueString(),
		"hours_types":          state.HoursType.ValueString(),
		"holiday_id":           state.HolidayID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("auto_receptionist_id"), req, resp)
}
