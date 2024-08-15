package user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

type resourceModel struct {
	ID                     types.String     `tfsdk:"id"`
	PhoneUserID            types.String     `tfsdk:"phone_user_id"`
	Email                  types.String     `tfsdk:"email"`
	FirstName              types.String     `tfsdk:"first_name"`
	LastName               types.String     `tfsdk:"last_name"`
	CallingPlans           types.Set        `tfsdk:"calling_plans"`
	SiteCode               types.String     `tfsdk:"site_code"`
	SiteName               types.String     `tfsdk:"site_name"`
	TemplateName           types.String     `tfsdk:"template_name"`
	ExtensionNumber        types.String     `tfsdk:"extension_number"`
	PhoneNumbers           types.Set        `tfsdk:"phone_numbers"`
	OutboundCallerID       types.String     `tfsdk:"outbound_caller_id"`
	SelectOutboundCallerID types.Bool       `tfsdk:"select_outbound_caller_id"`
	Sms                    smsResourceModel `tfsdk:"sms"`
}

type smsResourceModel struct {
	Enable                    types.Bool   `tfsdk:"enable"`
	InternationalSms          types.Bool   `tfsdk:"international_sms"`
	InternationalSmsCountries types.Set    `tfsdk:"international_sms_countries"`
	Locked                    types.Bool   `tfsdk:"locked"`
	LockedBy                  types.String `tfsdk:"locked_by"`
}

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	crud *userCrud
}

func (r userResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = fmt.Sprintf("%s_phone_user", request.ProviderTypeName)
}

func (r *userResource) Configure(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse) {
	panic("unimplemented")
}

func (r userResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	playRecordingBeepToneAttribute := schema.SingleNestedAttribute{
		MarkdownDescription: "Use this attribute to configure settings related to playing a recording beep tone.",
		Attributes: map[string]schema.Attribute{
			"enable": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to play a side tone beep for recorded users while recording. Only displayed when ad hoc call recording policy uses the new framework.",
			},
			"play_beep_volume": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int32{
					int32validator.OneOf(0, 20, 40, 60, 80, 100),
				},
				MarkdownDescription: "The volume of the side tone beep. It displays only when `enable` is set to `true`. Allowed: `0`, `20`, `40`, `60`, `80`, `100`",
			},
			"play_beep_time_interval": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int32{
					int32validator.OneOf(5, 10, 15, 20, 25, 30, 60, 120),
				},
				MarkdownDescription: "The beep time interval in seconds. It displays only when `enable` is `true`. Allowed: `5`, `10`, `15`, `20`, `25`, `30`, `60`, `120`",
			},
			"play_beep_member": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("allMember", "recordingSide"),
				},
				MarkdownDescription: "The beep sides. It displays only when `enable` is `true`. Allowed: `allMember`, `recordingSide`",
			},
		},
	}

	response.Schema = schema.Schema{
		MarkdownDescription: `Manages a user within Zoom Phone.

## API Permissions

The following API permissions are required in order to use this resource.

This resource requires the` + "`phone:write:batch_users:admin`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Zoom user.",
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The user email. It ensures the users are active in your Zoom account.",
			},
			"emergency_address_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The emergency address ID.",
			},
			"extension_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The extension ID.",
			},
			"extension_number": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The extension number of the user. The number must be complete (i.e. site number + short extension).",
			},
			"first_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The user's first name. It ensures the users are active in your Zoom account.",
			},
			"last_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The user's last name. It ensures the users are active in your Zoom account.",
			},
			"phone_user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Phone user.",
			},
			"policy": schema.SingleNestedAttribute{
				MarkdownDescription: "Use this attribute to configure settings related to the user's policy.",
				Attributes: map[string]schema.Attribute{
					"ad_hoc_call_recording": schema.SingleNestedAttribute{
						MarkdownDescription: "Use this attribute to configure settings related to ad hoc call recording.",
						Attributes: map[string]schema.Attribute{
							"enable": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether the current extension can record and save calls to the cloud.",
							},
							"locked": schema.BoolAttribute{
								Computed:            true,
								MarkdownDescription: "Whether the senior administrator allow users to modify the current settings.",
							},
							"locked_by": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Which level of administrator prohibits the modification of the current settings. Allowed: `account`, `user_group`, `site`",
							},
							"play_recording_beep_tone": playRecordingBeepToneAttribute,
							"recording_start_prompt": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether a prompt plays to call participants when the recording has started.",
							},
							"recording_transcription": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether call recording transcription is enabled.",
							},
						},
					},
					"auto_call_recording": schema.SingleNestedAttribute{
						MarkdownDescription: "Use this attribute to configure settings related to auto call recording.",
						Attributes: map[string]schema.Attribute{
							"allow_stop_resume_recording": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether the stop of and resuming of automatic call recording is enabled.",
							},
							"disconnect_on_recording_failure": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether a call disconnects when there is an issue with automatic call recording and the call cannot reconnect after five seconds. This does not include emergency calls.",
							},
							"enable": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether automatic call recording is enabled.",
							},
							"locked": schema.BoolAttribute{
								Computed:            true,
								MarkdownDescription: "Whether the senior administrator allow users to modify the current settings.",
							},
							"locked_by": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Which level of administrator prohibits the modification of the current settings. Allowed: `account`, `user_group`, `site`",
							},
							"play_recording_beep_tone": playRecordingBeepToneAttribute,
							"recording_calls": schema.StringAttribute{
								Optional: true,
								Computed: true,
								Validators: []validator.String{
									stringvalidator.OneOf("inbound", "outbound", "both"),
								},
								MarkdownDescription: "The type of calls automatically recorded. Allowed: `inbound`, `outbound`, `both`",
							},
							"recording_explicit_consent": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether press 1 to provide recording consent is enabled.",
							},
							"recording_start_prompt": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether a prompt plays to call participants when the recording has started.",
							},
							"recording_transcription": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether call recording transcription is enabled.",
							},
						},
					},
					"call_overflow": schema.SingleNestedAttribute{
						MarkdownDescription: "Use this attribute to configure settings related to call overflow.",
						Attributes: map[string]schema.Attribute{
							"call_overflow_type": schema.Int32Attribute{
								Optional: true,
								Computed: true,
								Validators: []validator.Int32{
									int32validator.Between(1, 4),
								},
								Default:             int32default.StaticInt32(4),
								MarkdownDescription: "`1` - Low restriction (external numbers not allowed) `2` - Medium restriction (external numbers and external contacts not allowed) `3` - High restriction (external numbers, external contacts and internal extensions without inbound automatic call recording not allowed) `4` - No restriction",
							},
							"enable": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(true),
								MarkdownDescription: "Whether to allow user to forward calls to other numbers.",
							},
							"locked": schema.BoolAttribute{
								Computed:            true,
								MarkdownDescription: "Whether the senior administrator allow users to modify the current settings.",
							},
							"locked_by": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Which level of administrator prohibits the modification of the current settings. Allowed: `account`, `user_group`, `site`",
							},
							"modified": schema.BoolAttribute{
								Computed:            true,
								MarkdownDescription: "Whether the current settings have been modified. If modified, they can be reset (displayed when using the new policy framework).",
							},
						},
					},
					"call_park": schema.SingleNestedAttribute{
						MarkdownDescription: "Use this attribute to configure settings related to call park.",
						Attributes: map[string]schema.Attribute{
							"call_not_picked_up_action": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "The action when a parked call is not picked up. 100-Ring back to parker, 0-Forward to voicemail of the parker, 9-Disconnect, 50-Forward to another extension.",
							},
							"enable": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether to allow calls placed on hold to be resumed from another location using a retrieval code.",
							},
							"expiration_period": schema.Int32Attribute{
								Optional: true,
								Computed: true,
								Validators: []validator.Int32{
									int32validator.OneOf(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60),
								},
								MarkdownDescription: "A time limit for parked calls, unit minutes. After the expiration period ends, the retrieval code is no longer valid and a new code will be generated. Allowed: `1`, `2`, `3`, `4`, `5`, `6`, `7`, `8`, `9`, `10`, `15`, `20`, `25`, `30`, `35`, `40`, `45`, `50`, `55`, `60`",
							},
							"forward_to_extension_id": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "The extension ID.",
							},
							"locked": schema.BoolAttribute{
								Computed:            true,
								MarkdownDescription: "Whether the senior administrator allow users to modify the current settings.",
							},
							"locked_by": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Which level of administrator prohibits the modification of the current settings. Allowed: `account`, `user_group`, `site`",
							},
						},
					},
					"call_transferring": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Use this attribute to configure settings related to call transferring.",
						Attributes: map[string]schema.Attribute{
							"call_transferring_type": schema.Int32Attribute{
								Optional: true,
								Computed: true,
								Validators: []validator.Int32{
									int32validator.Between(1, 4),
								},
								Default:             int32default.StaticInt32(1),
								MarkdownDescription: "`1` - No restriction. `2` - Medium restriction (external numbers and external contacts not allowed). `3` - High restriction (external numbers, unrecorded external contacts, and internal extensions without inbound automatic recording not allowed). `4` - Low restriction (external numbers not allowed). Allowed: `1`, `2`, `3`, `4`",
							},
							"enable": schema.BoolAttribute{
								Optional: true,
								Computed: true,
								Default:  booldefault.StaticBool(true),
							},
						},
					},
					"delegation": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether the user can use [call delegation](https://support.zoom.us/hc/en-us/articles/360032881731-Setting-up-call-delegation-shared-lines-appearance-).",
					},
					"elevate_to_meeting": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Whether the user can elevate their phone calls to a meeting.",
					},
					"emergency_address_management":            schema.SingleNestedAttribute{},
					"call_handling_forwarding_to_other_users": schema.SingleNestedAttribute{},
					"hand_off_to_room":                        schema.SingleNestedAttribute{},
					"international_calling": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Whether the current extension can make international calls outside of their calling plan.",
					},
					"mobile_switch_to_carrier":  schema.SingleNestedAttribute{},
					"select_outbound_caller_id": schema.SingleNestedAttribute{},
					"sms": schema.SingleNestedAttribute{
						MarkdownDescription: "Use this attribute to configure settings related to SMS.",
						Attributes: map[string]schema.Attribute{
							"enable": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether the user can send and receive messages.",
							},
							"international_sms": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether the user can send and receive international messages.",
							},
							"international_sms_countries": schema.SetAttribute{
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "The country that can send and receive international messages. Supported values are [the country ISO codes](https://marketplace.zoom.us/docs/api-reference/other-references/abbreviation-lists#countries).",
							},
							"locked": schema.BoolAttribute{
								Computed:            true,
								MarkdownDescription: "Whether the senior administrator allows users to modify the current settings.",
							},
							"locked_by": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Which level of administrator prohibits modifying the current settings. Allowed: `account`, `user_group` and `site`",
							},
						},
					},
					"voicemail":                              schema.SingleNestedAttribute{},
					"voicemail_access_members":               schema.SetNestedAttribute{},
					"zoom_phone_on_mobile":                   schema.SingleNestedAttribute{},
					"personal_audio_library":                 schema.SingleNestedAttribute{},
					"voicemail_transcription":                schema.SingleNestedAttribute{},
					"voicemail_notification_by_email":        schema.SingleNestedAttribute{},
					"shared_voicemail_notification_by_email": schema.SingleNestedAttribute{},
					"check_voicemails_over_phone":            schema.SingleNestedAttribute{},
					"audio_intercom":                         schema.SingleNestedAttribute{},
					"e2e_encryption":                         schema.SingleNestedAttribute{},
				},
			},
			"site_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the [site](https://support.zoom.us/hc/en-us/articles/360020809672z) where the user should be moved or assigned.",
			},
		},
	}
}

func (r userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	diag := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	ret, err := r.crud.create(ctx, &createDto{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone user",
			err.Error(),
		)
		return
	}

}

func (r userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (r userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (r userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	//TODO implement me
	panic("implement me")
}
