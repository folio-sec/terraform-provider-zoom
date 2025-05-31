package emergencyaddress

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DataSource{}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	crud *crud
}

type dataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	AddressLine1 types.String `tfsdk:"address_line1"`
	AddressLine2 types.String `tfsdk:"address_line2"`
	City         types.String `tfsdk:"city"`
	Country      types.String `tfsdk:"country"`
	IsDefault    types.Bool   `tfsdk:"is_default"`
	SiteID       types.String `tfsdk:"site_id"`
	StateCode    types.String `tfsdk:"state_code"`
	Zip          types.String `tfsdk:"zip"`
	Status       types.String `tfsdk:"status"`
	Level        types.String `tfsdk:"level"`
	UserID       types.String `tfsdk:"user_id"`
}

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_emergency_address"
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	markdownSeparatorForList := "\n  "
	resp.Schema = schema.Schema{
		MarkdownDescription: `Fetches information about a specific Zoom Phone emergency address.

## API Permissions

The following API permissions are required in order to use this data source.
This data source requires the ` + strings.Join([]string{
			"`phone:read:emergency_address:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The emergency address ID.",
			},
			"address_line1": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The emergency address line 1.",
			},
			"address_line2": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The emergency address line 2.",
			},
			"city": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The emergency address city.",
			},
			"country": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The two-lettered country code (Alpha-2 code in ISO-3166 format) of the emergency address.",
			},
			"is_default": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates whether the emergency address is default or not.",
			},
			"site_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the site to which this emergency address belongs.",
			},
			"state_code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The emergency address state code.",
			},
			"zip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The emergency address zip code.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The emergency address verification status." + strings.Join([]string{
					"",
					"- `1`: Verification not required.",
					"- `2`: Unverified.",
					"- `3`: Verification requested.",
					"- `4`: Verified.",
					"- `5`: Rejected.",
					"- `6`: Verification failed.",
				}, markdownSeparatorForList),
			},
			"level": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The emergency address owner level." + strings.Join([]string{
					"",
					"- `0`: Account/Company-level emergency address.",
					"- `1`: User/Personal-level emergency address.",
					"- `2`: Unknown company/pending emergency address.",
				}, markdownSeparatorForList),
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User ID to which the personal emergency address belongs.",
			},
		},
	}
}

func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*shared.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *shared.ProviderData, got: %T", req.ProviderData),
		)
		return
	}

	d.crud = newCrud(providerData.PhoneClient)
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.crud.read(ctx, config.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read emergency address, got error: %s", err))
		return
	}

	state := dataSourceModel{
		ID:           result.id,
		AddressLine1: result.addressLine1,
		AddressLine2: result.addressLine2,
		City:         result.city,
		Country:      result.country,
		IsDefault:    result.isDefault,
		SiteID:       result.site.ID,
		StateCode:    result.stateCode,
		Zip:          result.zip,
		Status:       types.StringValue(fmt.Sprintf("%d", result.status.ValueInt32())),
		Level:        types.StringValue(fmt.Sprintf("%d", result.level.ValueInt32())),
	}

	if !result.owner.ID.IsNull() {
		state.UserID = result.owner.ID
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
