package site

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var (
	_ datasource.DataSource              = &tfDataSource{}
	_ datasource.DataSourceWithConfigure = &tfDataSource{}
)

func NewPhoneSiteDataSource() datasource.DataSource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_site"
}

func (d *tfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `A Zoom Phone site in a Zoom account.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:site:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The site ID is the unique identifier of the site.",
			},
			"country": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The country of the site.",
				Attributes: map[string]schema.Attribute{
					"code": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The two lettered country [code](https://developers.zoom.us/docs/api/references/abbreviations/).",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The name of the country.",
					},
				},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the site.",
			},
			"main_auto_receptionist": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The [main auto receptionist](https://support.zoom.us/hc/en-us/articles/360021121312#h_bc7ff1d5-0e6c-40cd-b889-62010cb98c57) for each site.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The auto receptionist ID.",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Display name of the [auto-receptionist](https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0061421) as main auto-receptionist for the site.",
					},
				},
			},
			"site_code": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The [site code](https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0069806).",
			},
			"short_extension": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The short extension of the phone site.",
				Attributes: map[string]schema.Attribute{
					"length": schema.Int32Attribute{
						Computed:            true,
						MarkdownDescription: "This setting specifies the length of short extension numbers for the site. The value must be between 1 and 6., Default is `3`.",
					},
				},
			},
			"sip_zone": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "When you select a SIP zone nearest to your site, it might help reduce latency and improve call quality.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The SIP zone ID.",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The SIP zone name.",
					},
				},
			},
			"caller_id_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When an outbound call uses a number as the caller ID, the caller ID name and the number display to the called party. The caller ID name can be up to 15 characters. The user can reset the caller ID name by setting it to empty string.",
			},
			"level": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The level of the site.",
			},
			"india_state_code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The India site's state code. This field only applies to India based accounts.",
			},
			"india_city": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The India site's city. This field only applies to India based accounts.",
			},
			"india_sdca_npa": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The India site's Short Distance Calling Area (sdca) Numbering Plan Area (npa). This field is linked to the 'state_code' field. This field only applies to India based accounts.",
			},
			"india_entity_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When select the Indian sip zone, then need to set the entity name. This field only applies to India based accounts.",
			},
		},
	}
}

type dataSourceModel struct {
	ID                   types.String                         `tfsdk:"id"`
	Name                 types.String                         `tfsdk:"name"`
	Country              *dataSourceModelCountry              `tfsdk:"country"`
	MainAutoReceptionist *dataSourceModelMainAutoReceptionist `tfsdk:"main_auto_receptionist"`
	SiteCode             types.Int32                          `tfsdk:"site_code"`
	ShortExtension       *dataSourceModelShortExtension       `tfsdk:"short_extension"`
	SipZone              *dataSourceModelSipZone              `tfsdk:"sip_zone"`
	CallerIDName         types.String                         `tfsdk:"caller_id_name"`
	Level                types.String                         `tfsdk:"level"`
	IndiaStateCode       types.String                         `tfsdk:"india_state_code"`
	IndiaCity            types.String                         `tfsdk:"india_city"`
	IndiaSdcaNpa         types.String                         `tfsdk:"india_sdca_npa"`
	IndiaEntityName      types.String                         `tfsdk:"india_entity_name"`
}

type dataSourceModelCountry struct {
	Code types.String `tfsdk:"code"`
	Name types.String `tfsdk:"name"`
}

type dataSourceModelMainAutoReceptionist struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type dataSourceModelShortExtension struct {
	Length types.Int32 `tfsdk:"length"`
}

type dataSourceModelSipZone struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *tfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dto, err := d.crud.read(ctx, data.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone site", err.Error())
		return
	}
	if dto == nil {
		resp.Diagnostics.AddError(
			"Phone site not found",
			fmt.Sprintf("Phone site with ID %s not found", data.ID.ValueString()),
		)
		return
	}

	tflog.Info(ctx, "read phone site", map[string]interface{}{
		"id": dto.id.ValueString(),
	})

	output := dataSourceModel{
		ID:   dto.id,
		Name: dto.name,
		Country: &dataSourceModelCountry{
			Code: dto.country.code,
			Name: dto.country.name,
		},
		MainAutoReceptionist: &dataSourceModelMainAutoReceptionist{
			ID:   dto.mainAutoReceptionist.id,
			Name: dto.mainAutoReceptionist.name,
		},
		SiteCode: dto.siteCode,
		ShortExtension: &dataSourceModelShortExtension{
			Length: dto.shortExtensionLength,
		},
		SipZone: &dataSourceModelSipZone{
			ID:   dto.sipZone.id,
			Name: dto.sipZone.name,
		},
		CallerIDName:    dto.callerIDName,
		Level:           dto.level,
		IndiaStateCode:  lo.Ternary(dto.indiaStateCode.ValueString() == "", types.StringNull(), dto.indiaStateCode),
		IndiaCity:       lo.Ternary(dto.indiaCity.ValueString() == "", types.StringNull(), dto.indiaCity),
		IndiaSdcaNpa:    lo.Ternary(dto.indiaSdcaNpa.ValueString() == "", types.StringNull(), dto.indiaSdcaNpa),
		IndiaEntityName: lo.Ternary(dto.indiaEntityName.ValueString() == "", types.StringNull(), dto.indiaEntityName),
	}
	diags := resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
