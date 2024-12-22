package sharedlinegroupgroup

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

func NewPhoneSharedLineGroupDataSource() datasource.DataSource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_shared_line_group"
}

func (d *tfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `A [shared line group](https://support.zoom.us/hc/en-us/articles/360038850792) allows Zoom Phone admins to share a phone number and extension with a group of phone users or common areas. This gives members of the shared line group access to the group's direct phone number and voicemail.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:shared_line_group:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The unique identifier of the shared line group.",
			},
			"display_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name to identify the shared line group.",
			},
			"extension_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Extension ID.",
			},
			"extension_number": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Extension number of the shared line group.",
			},
			"primary_number": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `If you have multiple direct phone numbers assigned to the shared line group, this is the primary number selected for desk phones.
The primary number shares the same line as the extension number. This means if a caller is routed to the shared line group through an auto receptionist, the line associated with the primary number will be used.`,
			},
			"site_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the [site](https://support.zoom.us/hc/en-us/articles/360020809672-Managing-Multiple-Sites) where the shared line group is assigned.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `The status of the shared line group.
  - Allowed: active┃inactive`,
			},
		},
	}
}

type dataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	DisplayName     types.String `tfsdk:"display_name"`
	ExtensionID     types.String `tfsdk:"extension_id"`
	ExtensionNumber types.Int64  `tfsdk:"extension_number"`
	PrimaryNumber   types.String `tfsdk:"primary_number"`
	SiteID          types.String `tfsdk:"site_id"`
	Status          types.String `tfsdk:"status"`
}

func (d *tfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dto, err := d.crud.read(ctx, data.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone shared line group", err.Error())
		return
	}

	tflog.Info(ctx, "read phone shared line group", map[string]interface{}{
		"shared_line_group_id": dto.sharedLineGroupID.ValueString(),
	})

	siteID := types.StringNull()
	if dto.site != nil {
		siteID = dto.site.id
	}
	output := dataSourceModel{
		ID:              dto.sharedLineGroupID,
		ExtensionID:     dto.extensionID,
		ExtensionNumber: dto.extensionNumber,
		DisplayName:     dto.displayName,
		PrimaryNumber:   dto.primaryNumber,
		SiteID:          siteID,
		Status:          dto.status,
	}
	diags := resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
