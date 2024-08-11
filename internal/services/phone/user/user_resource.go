package user

import (
	"context"
	"fmt"
	"strings"

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

type userResourceModel struct {
	ID                     types.String         `tfsdk:"id"`
	PhoneUserID            types.String         `tfsdk:"phone_user_id"`
	Email                  types.String         `tfsdk:"email"`
	FirstName              types.String         `tfsdk:"first_name"`
	LastName               types.String         `tfsdk:"last_name"`
	CallingPlans           types.Set            `tfsdk:"calling_plans"`
	SiteCode               types.String         `tfsdk:"site_code"`
	SiteName               types.String         `tfsdk:"site_name"`
	TemplateName           types.String         `tfsdk:"template_name"`
	ExtensionNumber        types.String         `tfsdk:"extension_number"`
	PhoneNumbers           types.Set            `tfsdk:"phone_numbers"`
	OutboundCallerID       types.String         `tfsdk:"outbound_caller_id"`
	SelectOutboundCallerID types.Bool           `tfsdk:"select_outbound_caller_id"`
	Sms                    userSmsResourceModel `tfsdk:"sms"`
}

type userSmsResourceModel struct {
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
	playRecordingBeepToneBlock := schema.SingleNestedBlock{
		MarkdownDescription: "Use this block to configure settings related to playing a recording beep tone.",
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
			"phone_user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Phone user.",
			},
			"first_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The user's first name. It ensures the users are active in your Zoom account.",
			},
			"last_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The user's last name. It ensures the users are active in your Zoom account.",
			},
		},
		Blocks: map[string]schema.Block{
			"policy": schema.SingleNestedBlock{
				MarkdownDescription: "Use this block to configure settings related to the user's policy.",
				Blocks: map[string]schema.Block{
					"ad_hoc_call_recording": schema.SingleNestedBlock{
						MarkdownDescription: "Use this block to configure settings related to ad hoc call recording.",
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
							"reset": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether the user's ad hoc recording reset option will use the phone site's settings.",
							},
						},
						Blocks: map[string]schema.Block{
							"play_recording_beep_tone": playRecordingBeepToneBlock,
						},
					},
					"auto_call_recording": schema.SingleNestedBlock{
						MarkdownDescription: "Use this block to configure settings related to auto call recording.",
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
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether press 1 to provide recording consent is enabled.",
							},
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
						Blocks: map[string]schema.Block{
							"play_recording_beep_tone": playRecordingBeepToneBlock,
						},
					},
					"call_overflow": schema.SingleNestedBlock{
						MarkdownDescription: "Use this block to configure settings related to call overflow.",
						Attributes: map[string]schema.Attribute{
							"call_overflow_type": schema.Int32Attribute{
								Optional: true,
								Computed: true,
								Default:  int32default.StaticInt32(0),
								Validators: []validator.Int32{
									int32validator.Between(1, 4),
								},
								MarkdownDescription: "`1` - Low restriction (external numbers not allowed) `2` - Medium restriction (external numbers and external contacts not allowed) `3` - High restriction (external numbers, external contacts and internal extensions without inbound automatic call recording not allowed) `4` - No restriction",
							},
							"enable": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether to allow user to forward calls to other numbers.",
							},
							"reset": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Whether the current settings will use the phone site's settings (applicable if the current settings are using the new policy framework).",
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
								MarkdownDescription: "Whether the current settings have been modified. If modified, they can be reset (displayed when using the new policy framework)."
							},
						},
					},
					"call_park": schema.SingleNestedBlock{
						MarkdownDescription: "Use this block to configure settings related to call park.",
						Attributes: map[string]schema.Attribute{
							"call_not_picked_up_action": schema.Int64Attribute{
								Optional: true,
								Computed: true,
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
				},
			},
			"sms": schema.SingleNestedBlock{
				MarkdownDescription: "Use this block to configure settings related to SMS.",
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
		},
	}
}

func (r userResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r userResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (r userResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r userResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (r userResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	//TODO implement me
	panic("implement me")
}
