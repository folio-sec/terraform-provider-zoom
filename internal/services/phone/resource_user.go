package phone

import (
	"context"
	"fmt"
	"github.com/folio-sec/terraform-provider-zoom/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

type userResource struct {
	providerData *provider.Data
}

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResourceModel struct {
	ID                     types.String          `tfsdk:"id"`
	PhoneUserID            types.String          `tfsdk:"phone_user_id"`
	Email                  types.String          `tfsdk:"email"`
	FirstName              *types.String         `tfsdk:"first_name"`
	LastName               *types.String         `tfsdk:"last_name"`
	CallingPlans           types.Set             `tfsdk:"calling_plans"`
	SiteCode               *types.String         `tfsdk:"site_code"`
	SiteName               *types.String         `tfsdk:"site_name"`
	TemplateName           *types.String         `tfsdk:"template_name"`
	ExtensionNumber        types.String          `tfsdk:"extension_number"`
	PhoneNumbers           *types.Set            `tfsdk:"phone_numbers"`
	OutboundCallerID       *types.String         `tfsdk:"outbound_caller_id"`
	SelectOutboundCallerID *types.Bool           `tfsdk:"select_outbound_caller_id"`
	SMS                    *userSMSResourceModel `tfsdk:"sms"`
}

type userSMSResourceModel struct {
	Enable                    types.Bool   `tfsdk:"enable"`
	InternationalSMS          types.Bool   `tfsdk:"international_sms"`
	InternationalSMSCountries types.Set    `tfsdk:"international_sms_countries"`
	Locked                    types.Bool   `tfsdk:"locked"`
	LockedBy                  types.String `tfsdk:"locked_by"`
}

func (r userResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = fmt.Sprintf("%s_phone_user", request.ProviderTypeName)
}

func (r userResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"phone_user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Phone user.",
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The user email. It ensures the users are active in your Zoom account.",
			},
			"first_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The user's first name. It ensures the users are active in your Zoom account.",
			},
			"last_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The user's last name. It ensures the users are active in your Zoom account.",
			},
			"calling_plans": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Default:             setdefault.StaticValue(types.Set{}),
				MarkdownDescription: "The user's last name. It ensures the users are active in your Zoom account.",
			},
		},
		Blocks: map[string]schema.Block{
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
						MarkdownDescription: "Which level of administrator prohibits modifying the current settings. Allowed: `account`, `user_group` and `site`.",
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
