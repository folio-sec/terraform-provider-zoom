package callhandling

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var (
	_ resource.Resource                = &tfClosedHoursResource{}
	_ resource.ResourceWithConfigure   = &tfClosedHoursResource{}
	_ resource.ResourceWithImportState = &tfClosedHoursResource{}
)

func NewPhoneCallHandlingClosedHoursResource() resource.Resource {
	return &tfClosedHoursResource{}
}

type tfClosedHoursResource struct {
	crud *crud
}

func (r *tfClosedHoursResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tfClosedHoursResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_call_handling_closed_hours"
}

func (r *tfClosedHoursResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Call handling settings allow you to control how your system routes calls during closed hours.
For more information, read our [Call Handling API guide](https://developers.zoom.us/docs/zoom-phone/call-handling/) or Zoom support article [Customizing call handling settings](https://support.zoom.us/hc/en-us/articles/360059966372-Customizing-call-handling-settings).

## NOTE
This resource is depends on ` + "`zoom_phone_call_handling_business_hours`" + `. Please set business hours type = 2 (Custom hours).

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:call_handling_setting:admin`",
			"`phone:write:call_handling_setting:admin`",
			"`phone:update:call_handling_setting:admin`",
			"`phone:delete:call_handling_setting:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"extension_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Extension ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"call_handling": schema.SingleNestedAttribute{
				Required: true,
				MarkdownDescription: `The call handling settings.
  - NOTE: some fields doesn't return from zoom api, so please ignore_changes for these fields.
`,
				Attributes: map[string]schema.Attribute{
					"call_not_answer_action": schema.Int32Attribute{
						Optional: true,
						MarkdownDescription: `The action to take when a call is not answered:
  - 1 — Forward to a voicemail.
  - 2 — Forward to the user.
  - 4 — Forward to the common area.
  - 6 — Forward to the auto receptionist.
  - 7 — Forward to a call queue.
  - 8 — Forward to a shared line group.
  - 9 — Forward to an external contact.
  - 10 - Forward to a phone number.
  - 11 — Disconnect.
  - 12 — Play a message, then disconnect.
  - 13 - Forward to a message.
  - 14 - Forward to an interactive voice response (IVR).`,
						Validators: []validator.Int32{
							int32validator.OneOf(1, 2, 4, 6, 7, 8, 9, 10, 11, 12, 13, 14),
						},
					},
					"forward_to_extension_id": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: `The forwarding extension ID that's required only when call_not_answer_action setting is set to:
  - 2 - Forward to the user.
  - 4 - Forward to the common area.
  - 6 - Forward to the auto receptionist.
  - 7 - Forward to a call queue.
  - 8 - Forward to a shared line group.
  - 9 - forward to an external contact.`,
					},
					"busy_on_another_call_action": schema.Int32Attribute{
						Optional: true,
						MarkdownDescription: `The action to take when the user is busy on another call:
  - 1 — Forward to a voicemail.
  - 2 — Forward to the user.
  - 4 — Forward to the common area.
  - 6 — Forward to the auto receptionist.
  - 7 — Forward to a call queue.
  - 8 — Forward to a shared line group.
  - 9 — Forward to an external contact.
  - 10 - Forward to a phone number.
  - 12 — Play a message, then disconnect.
  - 21 — Call waiting.
  - 22 — Play a busy signal.`,
						Validators: []validator.Int32{
							int32validator.OneOf(1, 2, 4, 6, 7, 8, 9, 10, 12, 21, 22),
						},
					},
					"busy_forward_to_extension_id": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: `The forwarding extension ID that's required only when busy_on_another_call_action setting is set to:
  - 2 - Forward to the user.
  - 4 - Forward to the common area.
  - 6 - Forward to the auto receptionist.
  - 7 - Forward to a call queue.
  - 8 - Forward to a shared line group.
  - 9 - forward to an external contact.`,
					},
					"allow_callers_check_voicemail": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to allow the callers to check voicemails over a phone. It's required only when the call_not_answer_action setting is set to 1 (Forward to a voicemail).",
					},
					"busy_require_press1_before_connecting": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "When one is busy on another call, the receiver needs to press 1 before connecting the call for it to be forwarded to an external contact or a number. This option ensures that forwarded calls won't reach the voicemail box for the external contact or a number.",
					},
					"unanswered_require_press1_before_connecting": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "When a call is unanswered, press 1 before connecting the call to forward to an external contact or a number. This option ensures that forwarded calls won't reach the voicemail box for the external contact or a number.",
					},
					"overflow_play_callee_voicemail_greeting": schema.BoolAttribute{
						Optional: true,
						MarkdownDescription: `Whether to play the callee's voicemail greeting when the caller reaches the end of the forwarding sequence. It displays when call_not_answer_action is set to:
  - 2 - Forward to the user
  - 4 - Forward to the common area
  - 6 - Forward to the auto receptionist
  - 7 - Forward to a call queue
  - 8 - Forward to a shared line group
  - 9 - Forward to an external contact
  - 10 - Forward to an external number.`,
					},
					"play_callee_voicemail_greeting": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to play callee's voicemail greeting when the caller reaches the end of forwarding sequence. It displays when `busy_on_another_call_action` action or `call_not_answer_action` is set to `1` - Forward to a voicemail.",
					},
					"busy_play_callee_voicemail_greeting": schema.BoolAttribute{
						Optional: true,
						MarkdownDescription: `Whether to play callee's voicemail greeting when the caller reaches the end of the forwarding sequence. It displays when busy_on_another_call_action action is set to
  - 2 - Forward to the user
  - 4 - Forward to the common area
  - 6 - Forward to the auto receptionist
  - 7 - Forward to a call queue
  - 8 - Forward to a shared line group
  - 9 - Forward to an external contact
  - 10 - Forward to an external number.`,
					},
					"phone_number": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The extension's phone number or forward to an external number in [E.164](https://en.wikipedia.org/wiki/E.164) format format. It's required when `call_not_answer_action` action is set to `10` - Forward to an external number.",
					},
					"phone_number_description": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "(Optional) This field forwards to an external number description. Add this field when `call_not_answer_action` is set to `10` - Forward to an external number.",
					},
					"busy_phone_number": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The extension's phone number or forward to an external number in [E.164](https://en.wikipedia.org/wiki/E.164) format format. It sets when `busy_on_another_call_action` action is set to `10` - Forward to an external number.",
					},
					"busy_phone_number_description": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "This field forwards to an external number description (optional). It sets when `busy_on_another_call_action` action is set to `10` - Forward to an external number.",
					},
					"connect_to_operator": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to allow callers to reach an operator. It's required only when the `call_not_answer_action` or `busy_on_another_call_action` is set to 1 (Forward to a voicemail).",
					},
					"greeting_prompt_id": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: `The greeting audio prompt ID.
  - Options: empty char - default and 0 - disable
  - This is only required for the Call Queue or Auto Receptionist call_handling sub-setting.`,
					},
					"max_wait_time": schema.Int32Attribute{
						Optional: true,
						MarkdownDescription: `The maximum wait time, in seconds.
  - for simultaneous ring mode or the ring duration for each device for sequential ring mode: 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60.
  - Specify how long a caller will wait in the queue. Once the wait time is exceeded, the caller will be rerouted based on the overflow option for Call Queue: 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 120, 180, 240, 300, 600, 900, 1200, 1500, 1800.
  - This is only required for the call_handling sub-setting.
`,
						Validators: []validator.Int32{
							int32validator.OneOf(10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 120, 180, 240, 300, 600, 900, 1200, 1500, 1800),
						},
					},
					"operator_extension_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The extension ID of the operator to whom the call is being forwarded. It's required only when `call_not_answer_action` is set to `1` (Forward to a voicemail) and `connect_to_operator` is set to true.",
					},
					"ring_mode": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: `The call handling ring mode:
  - simultaneous
  - sequential. For user closed hours, ring_mode needs to be set with max_wait_time.`,
						Validators: []validator.String{
							stringvalidator.OneOf("simultaneous", "sequential"),
						},
					},
				},
			},
			"call_forwarding": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "The call forwarding settings.",
				Attributes: map[string]schema.Attribute{
					"require_press_1_before_connecting": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "When a call is forwarded to a personal phone number, whether the user must press \"1\" before the call connects. Enable this option to ensure missed calls do not reach to your personal voicemail. It's required for the `call_forwarding` sub-setting. Press 1 is always enabled and is required for callQueue type extension calls.",
					},
					"enable_zoom_mobile_apps": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to enable Zoom Mobile Apps call forwarding",
					},
					"enable_zoom_desktop_apps": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to enable Zoom Desktop Apps call forwarding",
					},
					"enable_zoom_phone_appliance_apps": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether to enable Zoom Phone Appliance Apps call forwarding",
					},
					"settings": schema.SetNestedAttribute{
						Optional:            true,
						MarkdownDescription: "The call forwarding settings. It's only required for the `call_forwarding` sub-setting.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The call forwarding's ID.",
								},
								"description": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "The external phone number's description.",
								},
								"enable": schema.BoolAttribute{
									Optional:            true,
									MarkdownDescription: "Whether to receive a call.",
								},
								"phone_number": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "The external phone number in [E.164](https://en.wikipedia.org/wiki/E.164) format format.",
								},
							},
						},
					},
				},
			},
		},
	}
}

type closedHoursResourceModel struct {
	ExtensionID    types.String                            `tfsdk:"extension_id"`
	CallHandling   *closedHoursResourceModelCallHandling   `tfsdk:"call_handling"`
	CallForwarding *closedHoursResourceModelCallForwarding `tfsdk:"call_forwarding"`
}

type closedHoursResourceModelCallHandling struct {
	CallNotAnswerAction                     types.Int32  `tfsdk:"call_not_answer_action"`
	ForwardToExtensionID                    types.String `tfsdk:"forward_to_extension_id"`
	BusyOnAnotherCallAction                 types.Int32  `tfsdk:"busy_on_another_call_action"`
	BusyForwardToExtensionID                types.String `tfsdk:"busy_forward_to_extension_id"`
	AllowCallersCheckVoicemail              types.Bool   `tfsdk:"allow_callers_check_voicemail"`
	BusyRequirePress1BeforeConnecting       types.Bool   `tfsdk:"busy_require_press1_before_connecting"`
	UnAnsweredRequirePress1BeforeConnecting types.Bool   `tfsdk:"unanswered_require_press1_before_connecting"`
	OverflowPlayCalleeVoicemailGreeting     types.Bool   `tfsdk:"overflow_play_callee_voicemail_greeting"`
	PlayCalleeVoicemailGreeting             types.Bool   `tfsdk:"play_callee_voicemail_greeting"`
	BusyPlayCalleeVoicemailGreeting         types.Bool   `tfsdk:"busy_play_callee_voicemail_greeting"`
	PhoneNumber                             types.String `tfsdk:"phone_number"`
	PhoneNumberDescription                  types.String `tfsdk:"phone_number_description"`
	BusyPhoneNumber                         types.String `tfsdk:"busy_phone_number"`
	BusyPhoneNumberDescription              types.String `tfsdk:"busy_phone_number_description"`
	ConnectToOperator                       types.Bool   `tfsdk:"connect_to_operator"`
	GreetingPromptID                        types.String `tfsdk:"greeting_prompt_id"`
	MaxWaitTime                             types.Int32  `tfsdk:"max_wait_time"`
	OperatorExtensionID                     types.String `tfsdk:"operator_extension_id"`
	RingMode                                types.String `tfsdk:"ring_mode"`
}

type closedHoursResourceModelCallForwarding struct {
	RequirePress1BeforeConnecting types.Bool                                       `tfsdk:"require_press_1_before_connecting"`
	EnableZoomMobileApps          types.Bool                                       `tfsdk:"enable_zoom_mobile_apps"`
	EnableZoomDesktopApps         types.Bool                                       `tfsdk:"enable_zoom_desktop_apps"`
	EnableZoomPhoneApplianceApps  types.Bool                                       `tfsdk:"enable_zoom_phone_appliance_apps"`
	Settings                      []*closedHoursResourceModelCallForwardingSetting `tfsdk:"settings"`
}

type closedHoursResourceModelCallForwardingSetting struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Enable      types.Bool   `tfsdk:"enable"`
	PhoneNumber types.String `tfsdk:"phone_number"`
}

func (r *tfClosedHoursResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state closedHoursResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.read(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone call handling", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfClosedHoursResource) read(ctx context.Context, plan *closedHoursResourceModel) (*closedHoursResourceModel, error) {
	dto, err := r.crud.readClosedHours(ctx, plan.ExtensionID)
	if err != nil {
		return nil, fmt.Errorf("error read: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	callHandling := &closedHoursResourceModelCallHandling{
		CallNotAnswerAction:                     dto.callHandling.callNotAnswerAction,
		ForwardToExtensionID:                    dto.callHandling.forwardToExtensionID,
		BusyOnAnotherCallAction:                 dto.callHandling.busyOnAnotherCallAction,
		BusyForwardToExtensionID:                dto.callHandling.busyForwardToExtensionID,
		AllowCallersCheckVoicemail:              dto.callHandling.allowCallersCheckVoicemail,
		BusyRequirePress1BeforeConnecting:       dto.callHandling.busyRequirePress1BeforeConnecting,
		UnAnsweredRequirePress1BeforeConnecting: dto.callHandling.unAnsweredRequirePress1BeforeConnecting,
		OverflowPlayCalleeVoicemailGreeting:     dto.callHandling.overflowPlayCalleeVoicemailGreeting,
		PlayCalleeVoicemailGreeting:             dto.callHandling.playCalleeVoicemailGreeting,
		BusyPlayCalleeVoicemailGreeting:         dto.callHandling.busyPlayCalleeVoicemailGreeting,
		PhoneNumber:                             dto.callHandling.phoneNumber,
		PhoneNumberDescription:                  dto.callHandling.phoneNumberDescription,
		BusyPhoneNumber:                         dto.callHandling.busyPhoneNumber,
		BusyPhoneNumberDescription:              dto.callHandling.busyPhoneNumberDescription,
		ConnectToOperator:                       dto.callHandling.connectToOperator,
		GreetingPromptID:                        dto.callHandling.greetingPromptID,
		MaxWaitTime:                             dto.callHandling.maxWaitTime,
		OperatorExtensionID:                     dto.callHandling.operatorExtensionID,
		RingMode:                                dto.callHandling.ringMode,
	}
	callForwarding := lo.TernaryF(plan.CallForwarding != nil && dto.callForwarding != nil, func() *closedHoursResourceModelCallForwarding {
		return &closedHoursResourceModelCallForwarding{
			RequirePress1BeforeConnecting: dto.callForwarding.requirePress1BeforeConnecting,
			EnableZoomMobileApps:          dto.callForwarding.enableZoomMobileApps,
			EnableZoomDesktopApps:         dto.callForwarding.enableZoomDesktopApps,
			EnableZoomPhoneApplianceApps:  dto.callForwarding.enableZoomPhoneApplianceApps,
			Settings: lo.Map(dto.callForwarding.settings, func(item *readDtoCallForwardingSetting, index int) *closedHoursResourceModelCallForwardingSetting {
				return &closedHoursResourceModelCallForwardingSetting{
					ID:          item.id,
					Description: item.description,
					Enable:      item.enable,
					PhoneNumber: item.phoneNumber,
				}
			}),
		}
	}, func() *closedHoursResourceModelCallForwarding {
		return nil
	})
	return &closedHoursResourceModel{
		ExtensionID:    dto.extensionID,
		CallHandling:   callHandling,
		CallForwarding: callForwarding,
	}, nil
}

func (r *tfClosedHoursResource) sync(ctx context.Context, plan *closedHoursResourceModel, onDelete bool) error {
	// closed_hours sync handling like followings
	// 1. call_handling contains one setting and cannot delete it -> do PATCH
	// 2. call_forwarding may contain one setting and can delete it  -> do CREATE/PATCH/DELETE

	asis, err := r.read(ctx, plan)
	if err != nil {
		return err
	}

	// 0. Validate
	businessHours, err := r.crud.readBusinessHours(ctx, plan.ExtensionID)
	if err != nil {
		return err
	}
	if businessHours == nil {
		if onDelete {
			return nil
		}
		return fmt.Errorf("business_hours not found")
	}
	if businessHours.customHours.typ.ValueInt32() != 2 {
		// closed_hours can set after setting business_hours.custom_hours.type = 2(custom_hours)
		return fmt.Errorf("closed_hours can set after setting business_hours.custom_hours.type = 2, please add depends_on setting with business_hours")
	}

	// 1. PATCH call_handling
	patchCallHandling := &patchCallHandlingDto{
		extensionID: plan.ExtensionID,
		settingType: settingTypeClosedHours,
		settings: &patchCallHandlingDtoSettings{
			callNotAnswerAction:                     plan.CallHandling.CallNotAnswerAction,
			forwardToExtensionID:                    plan.CallHandling.ForwardToExtensionID,
			busyOnAnotherCallAction:                 plan.CallHandling.BusyOnAnotherCallAction,
			busyForwardToExtensionID:                plan.CallHandling.BusyForwardToExtensionID,
			allowCallersCheckVoicemail:              plan.CallHandling.AllowCallersCheckVoicemail,
			busyRequirePress1BeforeConnecting:       plan.CallHandling.BusyRequirePress1BeforeConnecting,
			unAnsweredRequirePress1BeforeConnecting: plan.CallHandling.UnAnsweredRequirePress1BeforeConnecting,
			overflowPlayCalleeVoicemailGreeting:     plan.CallHandling.OverflowPlayCalleeVoicemailGreeting,
			playCalleeVoicemailGreeting:             plan.CallHandling.PlayCalleeVoicemailGreeting,
			busyPlayCalleeVoicemailGreeting:         plan.CallHandling.BusyPlayCalleeVoicemailGreeting,
			phoneNumber:                             plan.CallHandling.PhoneNumber,
			phoneNumberDescription:                  plan.CallHandling.PhoneNumberDescription,
			busyPhoneNumber:                         plan.CallHandling.BusyPhoneNumber,
			busyPhoneNumberDescription:              plan.CallHandling.BusyPhoneNumberDescription,
			connectToOperator:                       plan.CallHandling.ConnectToOperator,
			greetingPromptID:                        plan.CallHandling.GreetingPromptID,
			maxWaitTime:                             plan.CallHandling.MaxWaitTime,
			operatorExtensionID:                     plan.CallHandling.OperatorExtensionID,
			ringMode:                                plan.CallHandling.RingMode,
		},
	}
	if err = r.crud.patchCallHandling(ctx, patchCallHandling); err != nil {
		return err
	}

	// 2. CREATE/PATCH/DELETE call_forwarding
	// 2-1: PATCH existed call forwarding
	// 2-2: create new call forwarding
	// 2-3: delete unused call forwarding

	// 2-1: PATCH existed call forwarding
	if plan.CallForwarding != nil {
		var settings []*patchCallForwardingDtoSetting
		if plan.CallForwarding.Settings != nil {
			settings = lo.FilterMap(plan.CallForwarding.Settings, func(item *closedHoursResourceModelCallForwardingSetting, index int) (*patchCallForwardingDtoSetting, bool) {
				return &patchCallForwardingDtoSetting{
					id:          item.ID,
					description: item.Description,
					enable:      item.Enable,
					phoneNumber: item.PhoneNumber,
				}, item.ID.ValueString() != "" // patch only id is existed
			})
		}
		patchCallForwarding := &patchCallForwardingDto{
			extensionID:                   plan.ExtensionID,
			settingType:                   settingTypeClosedHours,
			requirePress1BeforeConnecting: plan.CallForwarding.RequirePress1BeforeConnecting,
			enableZoomMobileApps:          plan.CallForwarding.EnableZoomMobileApps,
			enableZoomDesktopApps:         plan.CallForwarding.EnableZoomDesktopApps,
			enableZoomPhoneApplianceApps:  plan.CallForwarding.EnableZoomPhoneApplianceApps,
			settings:                      settings,
		}
		if err = r.crud.patchCallForwarding(ctx, patchCallForwarding, onDelete); err != nil {
			return err
		}
	} else if asis.CallForwarding != nil {
		patchCallForwarding := &patchCallForwardingDto{
			extensionID:                   plan.ExtensionID,
			settingType:                   settingTypeClosedHours,
			requirePress1BeforeConnecting: types.BoolValue(false),
			enableZoomMobileApps:          types.BoolValue(true),
			enableZoomDesktopApps:         types.BoolValue(true),
			enableZoomPhoneApplianceApps:  types.BoolValue(true),
			settings:                      []*patchCallForwardingDtoSetting{},
		}
		if err = r.crud.patchCallForwarding(ctx, patchCallForwarding, onDelete); err != nil {
			return err
		}
	}

	// 2-2: create new call forwarding
	if plan.CallForwarding != nil {
		newCallForwardings := lo.Filter(plan.CallForwarding.Settings, func(item *closedHoursResourceModelCallForwardingSetting, index int) bool {
			return item.ID.ValueString() == ""
		})
		for _, newCallForwarding := range newCallForwardings {
			createAllForwardingDto := &createCallForwardingDto{
				extensionID: plan.ExtensionID,
				settingType: settingTypeClosedHours,
				description: newCallForwarding.Description,
				phoneNumber: newCallForwarding.PhoneNumber,
			}
			created, err := r.crud.createCallForwarding(ctx, createAllForwardingDto)
			if err != nil {
				return err
			}
			newCallForwarding.ID = created.callForwardingID
		}
	}

	// 2-3: delete unused call forwarding
	if asis.CallForwarding != nil {
		deleteCallForwardings := lo.Filter(asis.CallForwarding.Settings, func(item *closedHoursResourceModelCallForwardingSetting, index int) bool {
			if plan.CallForwarding == nil {
				return true
			}
			for _, planCallForwarding := range plan.CallForwarding.Settings {
				if item.ID == planCallForwarding.ID {
					return false
				}
			}
			return true
		})
		for _, setting := range deleteCallForwardings {
			deleteCallForwarding := &deleteCallForwardingDto{
				extensionID:      asis.ExtensionID,
				settingType:      settingTypeClosedHours,
				callForwardingID: setting.ID,
			}
			if err = r.crud.deleteCallForwarding(ctx, deleteCallForwarding); err != nil {
				return nil
			}
		}
	}

	return nil
}

func (r *tfClosedHoursResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan closedHoursResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.sync(ctx, &plan, false); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call handling",
			err.Error(),
		)
		return
	}

	// XXX zoom api doesn't return some fields after create/update, so just set plan with existed plan values
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfClosedHoursResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan closedHoursResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error updating phone call handling",
			"Error updating phone call handling",
		)
		return
	}

	if err := r.sync(ctx, &plan, false); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone call handling",
			fmt.Sprintf(
				"Could not update phone call handling %s, unexpected error: %s",
				plan.ExtensionID.ValueString(),
				err,
			),
		)
		return
	}

	// XXX zoom api doesn't return some fields after create/update, so just set plan with existed plan values
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfClosedHoursResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state closedHoursResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultModel := closedHoursResourceModel{
		ExtensionID:    state.ExtensionID,
		CallHandling:   &closedHoursResourceModelCallHandling{},
		CallForwarding: &closedHoursResourceModelCallForwarding{},
	}
	err := r.sync(ctx, &defaultModel, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone call handling",
			fmt.Sprintf(
				"Could not delete phone call handling %s, unexpected error: %s",
				state.ExtensionID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone call handling", map[string]interface{}{
		"extension_id": state.ExtensionID.ValueString(),
	})
}

func (r *tfClosedHoursResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("extension_id"), req, resp)
}
