package autoreceptionist

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &tfDataSource{}
	_ datasource.DataSourceWithConfigure = &tfDataSource{}
)

func NewPhoneAutoReceptionistDataSource() datasource.DataSource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_auto_receptionist"
}

func (d *tfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Auto receptionists answer calls with a personalized recording and routes calls to a phone user, call queue, common area, voicemail or an IVR system.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:auto_receptionist:admin`",
		}, ", "),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Auto receptionist ID. The unique identifier of the auto receptionist.",
			},
			"cost_center": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cost center name.",
			},
			"department": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Department name.",
			},
			"extension_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Extension ID.",
			},
			"extension_number": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Extension number of the auto receptionist.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the auto receptionist.",
			},
			"timezone": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "[Timezone](https://marketplace.zoom.us/docs/api-reference/other-references/abbreviation-lists#timezones) of the Auto Receptionist.",
			},
			"audio_prompt_language": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: `The language for all default audio prompts for the auto receptionist.`,
			},
		},
	}
}

type dataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	CostCenter          types.String `tfsdk:"cost_center"`
	Department          types.String `tfsdk:"department"`
	ExtensionID         types.String `tfsdk:"extension_id"`
	ExtensionNumber     types.Int64  `tfsdk:"extension_number"`
	Name                types.String `tfsdk:"name"`
	Timezone            types.String `tfsdk:"timezone"`
	AudioPromptLanguage types.String `tfsdk:"audio_prompt_language"`
}

func (d *tfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dto, err := d.crud.read(ctx, data.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone auto receptionist", err.Error())
		return
	}

	tflog.Info(ctx, "read phone auto receptionist", map[string]interface{}{
		"auto_receptionist_id": dto.autoReceptionistID.ValueString(),
	})

	output := dataSourceModel{
		ID:                  dto.autoReceptionistID,
		CostCenter:          dto.costCenter,
		Department:          dto.department,
		ExtensionID:         dto.extensionID,
		ExtensionNumber:     dto.extensionNumber,
		Name:                dto.name,
		Timezone:            dto.timezone,
		AudioPromptLanguage: dto.audioPromptLanguage,
	}
	diags := resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
