package callqueue

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

func NewPhoneCallQueueDataSource() datasource.DataSource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_call_queue"
}

func (d *tfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Call queues allow you to route incoming calls to a group of users. For instance, you can use call queues to route calls to various departments in your organization such as sales, engineering, billing, customer service etc.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:call_queue:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier of the Call Queue.",
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
				MarkdownDescription: "Extension number of the call queue.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the call queue.",
			},
			"site_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the [site](https://support.zoom.us/hc/en-us/articles/360020809672-Managing-Multiple-Sites) where the Call Queue is assigned.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `Status of the Call Queue.
  - Allowed: active┃inactive`,
			},
		},
	}
}

type dataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	CostCenter      types.String `tfsdk:"cost_center"`
	Department      types.String `tfsdk:"department"`
	ExtensionID     types.String `tfsdk:"extension_id"`
	ExtensionNumber types.Int64  `tfsdk:"extension_number"`
	Name            types.String `tfsdk:"name"`
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
		resp.Diagnostics.AddError("Error reading phone call queue", err.Error())
		return
	}

	tflog.Info(ctx, "read phone call queue", map[string]interface{}{
		"call_queue_id": dto.callQueueID.ValueString(),
	})

	siteID := types.StringNull()
	if dto.site != nil {
		siteID = dto.site.id
	}
	output := dataSourceModel{
		ID:              dto.callQueueID,
		CostCenter:      dto.costCenter,
		Department:      dto.department,
		ExtensionID:     dto.extensionID,
		ExtensionNumber: dto.extensionNumber,
		Name:            dto.name,
		SiteID:          siteID,
		Status:          dto.status,
	}
	diags := resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
