package phonenumbers

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var (
	_ datasource.DataSource              = &tfDataSource{}
	_ datasource.DataSourceWithConfigure = &tfDataSource{}
)

func NewPhonePhoneNumbersDataSource() datasource.DataSource {
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
	d.crud = newCrud(data.PhoneClient)
}

func (d *tfDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_phone_numbers"
}

func (d *tfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Zoom Phone numbers in a Zoom account.

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:list_numbers:admin`",
		}, ", "),
		Attributes: map[string]schema.Attribute{
			"filter": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: `The query response by number assignment. The value can be one of the following:
  - assigned: The number has been assigned to either a user, a call queue, an auto-receptionist, or a common area in an account.
  - unassigned: The number is not assigned to anyone.
  - all: Include both assigned and unassigned numbers in the response.
  - byoc: Include Bring Your Own Carrier (BYOC) numbers only in the response.`,
						Validators: []validator.String{
							stringvalidator.OneOf("assigned", "unassigned", "all", "byoc"),
						},
					},
					"extension_type": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: `The type of assignee to whom the number is assigned. The parameter can be set only if type parameter is set as assigned. The value can be one of the following:
  - Allowed: user ┃ callQueue ┃ autoReceptionist ┃ commonArea ┃ emergencyNumberPool ┃ companyLocation ┃ meetingService`,
						Validators: []validator.String{
							stringvalidator.OneOf("user", "callQueue", "autoReceptionist", "commonArea", "emergencyNumberPool", "companyLocation", "meetingService"),
						},
					},
					"number_type": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The type of number. Values can be one of the following: `toll`, `tollfree`.",
						Validators: []validator.String{
							stringvalidator.OneOf("toll", "tollfree"),
						},
					},
					"pending_numbers": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "This field includes or excludes pending numbers in the response.",
					},
					"site_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The unique identifier of the site. Use this query parameter if you have enabled multiple sites and would like to filter the response of this API call by a specific phone site. See [Managing multiple sites](https://support.zoom.us/hc/en-us/articles/360020809672-Managing-multiple-sites) or [Adding a site](https://support.zoom.us/hc/en-us/articles/360020809672-Managing-multiple-sites#h_05c88e35-1593-491f-b1a8-b7139a75dc15) for details.",
					},
				},
			},
			"phone_numbers": schema.ListNestedAttribute{
				Computed: true,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"assignee": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"extension_number": schema.Int64Attribute{
									Computed:            true,
									MarkdownDescription: "The extension number of the phone.",
								},
								"id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The unique identifier of the user to whom the number has been assigned.",
								},
								"name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The name of the user to whom the number has been assigned.",
								},
								"type": schema.StringAttribute{
									Computed: true,
									MarkdownDescription: `
This field indicates to whom the phone number belongs.
  - user: Number has been assigned to an existing phone user that allows them to receive calls through their extension number or direct phone number.
  - callQueue: Phone number has been assigned to a call queue.
  - autoReceptionist: Phone number has been assigned to an auto receptionist.
  - commonArea: Phone number has been assigned to a common area.
  - emergencyNumberPool
  - companyLocation
  - meetingService`,
								},
							},
						},
						"capability": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "The capability for the phone number, whether it can take incoming calls, make outgoing calls, or both. Values include `incoming`, `outgoing`, or both values.",
						},
						"carrier": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "This field displays when the `type` request parameter is `byoc`.",
							Attributes: map[string]schema.Attribute{
								"code": schema.Int32Attribute{
									Computed:            true,
									MarkdownDescription: "The carrier code.",
								},
								"name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The name of the carrier to which the phone number is assigned.",
								},
							},
						},
						"display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The display name for the phone number.",
						},
						"emergency_address": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "This field displays when the `type` request parameter is `byoc`.",
							Attributes: map[string]schema.Attribute{
								"address_line1": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The address Line 1 of the [emergency address](https://support.zoom.us/hc/en-us/articles/360021062871-Setting-an-Emergency-Address) that consists of the house number and street name.",
								},
								"address_line2": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Address Line 2 of the [emergency address](https://support.zoom.us/hc/en-us/articles/360021062871-Setting-an-Emergency-Address) that consists of building number, floor number, unit and so on.",
								},
								"city": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The city of the [emergency address](https://support.zoom.us/hc/en-us/articles/360021062871-Setting-an-Emergency-Address).",
								},
								"country": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The two-lettered country code (Aplha-2 code in ISO-3166 format) standard of the site's [emergency address](https://support.zoom.us/hc/en-us/articles/360021062871-Setting-an-Emergency-Address).",
								},
								"state_code": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The state code of the [emergency address](https://support.zoom.us/hc/en-us/articles/360021062871-Setting-an-Emergency-Address).",
								},
								"zip": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The zip code of the [emergency address](https://support.zoom.us/hc/en-us/articles/360021062871-Setting-an-Emergency-Address).",
								},
							},
						},
						"emergency_address_status": schema.Int32Attribute{
							Computed:            true,
							MarkdownDescription: "This field displays when the `type` request parameter is `byoc`. The emergency address status: `1`-carrier update required, `2`-confirmed.",
						},
						"emergency_address_update_time": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "This field displays when the `type` request parameter is `byoc`. The time of emergency address information update (format: 'yyyy-MM-ddThh:dd:ssZ').",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the phone number.",
						},
						"location": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The location (city, state and country) where the phone number is assigned.",
						},
						"number": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The phone number in E164 format.",
						},
						"number_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The type of number. Values can be one of the following: `toll`, `tollfree`.",
						},
						"sip_group": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "This field displays when the `type` request parameter is `byoc`.",
							Attributes: map[string]schema.Attribute{
								"display_name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The name of the SIP group.",
								},
								"id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The ID of the SIP group. See the **Creating SIP groups** section in [Creating a shared directory of external contacts](https://support.zoom.us/hc/en-us/articles/360037050092-Creating-a-shared-directory-of-external-contacts) for details.",
								},
							},
						},
						"site": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The target [site](https://support.zoom.us/hc/en-us/articles/360020809672-Managing-Multiple-Sites) in which the phone number was assigned. Sites allow you to organize the phone users in your organization. For example, you sites could be created based on different office locations.",
								},
								"name": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The name of the site where the phone number is assigned.",
								},
							},
						},
						"source": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The source of the phone number.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The status of the number.",
						},
					},
				},
			},
		},
	}
}

type dataSourceModel struct {
	Filter       *dataSourceModelFilter        `tfsdk:"filter"`
	PhoneNumbers []*dataSourceModelPhoneNumber `tfsdk:"phone_numbers"`
}

type dataSourceModelPhoneNumber struct {
	Assignee                   *dataSourceModelAssignee         `tfsdk:"assignee"`
	Capability                 []types.String                   `tfsdk:"capability"`
	Carrier                    *dataSourceModelCarrier          `tfsdk:"carrier"`
	DisplayName                types.String                     `tfsdk:"display_name"`
	EmergencyAddress           *dataSourceModelEmergencyAddress `tfsdk:"emergency_address"`
	EmergencyAddressStatus     types.Int32                      `tfsdk:"emergency_address_status"`
	EmergencyAddressUpdateTime types.String                     `tfsdk:"emergency_address_update_time"`
	ID                         types.String                     `tfsdk:"id"`
	Location                   types.String                     `tfsdk:"location"`
	Number                     types.String                     `tfsdk:"number"`
	NumberType                 types.String                     `tfsdk:"number_type"`
	SipGroup                   *dataSourceModelSipGroup         `tfsdk:"sip_group"`
	Site                       *dataSourceModelSite             `tfsdk:"site"`
	Source                     types.String                     `tfsdk:"source"`
	Status                     types.String                     `tfsdk:"status"`
}

type dataSourceModelFilter struct {
	Type           types.String `tfsdk:"type"`
	ExtensionType  types.String `tfsdk:"extension_type"`
	NumberType     types.String `tfsdk:"number_type"`
	PendingNumbers types.Bool   `tfsdk:"pending_numbers"`
	SiteID         types.String `tfsdk:"site_id"`
}

type dataSourceModelAssignee struct {
	ExtensionNumber types.Int64  `tfsdk:"extension_number"`
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
}

type dataSourceModelCarrier struct {
	Code types.Int32  `tfsdk:"code"`
	Name types.String `tfsdk:"name"`
}

type dataSourceModelEmergencyAddress struct {
	AddressLine1 types.String `tfsdk:"address_line1"`
	AddressLine2 types.String `tfsdk:"address_line2"`
	City         types.String `tfsdk:"city"`
	Country      types.String `tfsdk:"country"`
	StateCode    types.String `tfsdk:"state_code"`
	Zip          types.String `tfsdk:"zip"`
}

type dataSourceModelSipGroup struct {
	DisplayName types.String `tfsdk:"display_name"`
	ID          types.String `tfsdk:"id"`
}

type dataSourceModelSite struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *tfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	typ := types.StringNull()
	extensionType := types.StringNull()
	numberType := types.StringNull()
	pendingNumbers := types.BoolNull()
	siteID := types.StringNull()
	if data.Filter != nil {
		typ = data.Filter.Type
		extensionType = data.Filter.ExtensionType
		numberType = data.Filter.NumberType
		pendingNumbers = data.Filter.PendingNumbers
		siteID = data.Filter.SiteID
	}
	dto, err := d.crud.read(ctx, &readQueryDto{
		typ:            typ,
		extensionType:  extensionType,
		numberType:     numberType,
		pendingNumbers: pendingNumbers,
		siteID:         siteID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone numbers", err.Error())
		return
	}

	tflog.Info(ctx, "read phone numbers")

	var filter *dataSourceModelFilter
	if data.Filter != nil {
		filter = &dataSourceModelFilter{
			Type:           data.Filter.Type,
			ExtensionType:  data.Filter.ExtensionType,
			NumberType:     data.Filter.NumberType,
			PendingNumbers: data.Filter.PendingNumbers,
			SiteID:         data.Filter.SiteID,
		}
	}
	output := dataSourceModel{
		Filter: filter,
		PhoneNumbers: lo.Map(dto.phoneNumbers, func(item *readDtoPhoneNumber, _index int) *dataSourceModelPhoneNumber {
			var assignee *dataSourceModelAssignee
			if item.assignee != nil {
				assignee = &dataSourceModelAssignee{
					ExtensionNumber: item.assignee.extensionNumber,
					ID:              item.assignee.id,
					Name:            item.assignee.name,
					Type:            item.assignee.typ,
				}
			}
			var carrier *dataSourceModelCarrier
			if item.carrier != nil {
				carrier = &dataSourceModelCarrier{
					Code: item.carrier.code,
					Name: item.carrier.name,
				}
			}
			var emergencyAddress *dataSourceModelEmergencyAddress
			if item.emergencyAddress != nil {
				emergencyAddress = &dataSourceModelEmergencyAddress{
					AddressLine1: item.emergencyAddress.addressLine1,
					AddressLine2: item.emergencyAddress.addressLine2,
					City:         item.emergencyAddress.city,
					Country:      item.emergencyAddress.country,
					StateCode:    item.emergencyAddress.stateCode,
					Zip:          item.emergencyAddress.zip,
				}
			}
			var sipGroup *dataSourceModelSipGroup
			if item.sipGroup != nil {
				sipGroup = &dataSourceModelSipGroup{
					DisplayName: item.sipGroup.displayName,
					ID:          item.sipGroup.id,
				}
			}
			var site *dataSourceModelSite
			if item.site != nil {
				site = &dataSourceModelSite{
					ID:   item.site.id,
					Name: item.site.name,
				}
			}
			return &dataSourceModelPhoneNumber{
				Assignee:                   assignee,
				Capability:                 item.capability,
				Carrier:                    carrier,
				DisplayName:                item.displayName,
				EmergencyAddress:           emergencyAddress,
				EmergencyAddressStatus:     item.emergencyAddressStatus,
				EmergencyAddressUpdateTime: item.emergencyAddressUpdateTime,
				ID:                         item.id,
				Location:                   item.location,
				Number:                     item.number,
				NumberType:                 item.numberType,
				SipGroup:                   sipGroup,
				Site:                       site,
				Source:                     item.source,
				Status:                     item.status,
			}
		}),
	}

	diags := resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
