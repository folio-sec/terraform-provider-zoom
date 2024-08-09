package autoreceptionist

import (
	"context"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &phoneAutoReceptionistDataSource{}
	_ datasource.DataSourceWithConfigure = &phoneAutoReceptionistDataSource{}
)

func NewPhoneAutoReceptionistDataSource() datasource.DataSource {
	return &phoneAutoReceptionistDataSource{}
}

type phoneAutoReceptionistDataSource struct {
	crud *PhoneAutoReceptionistCrud
}

func (d *phoneAutoReceptionistDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_auto_receptionist"
}

func (d *phoneAutoReceptionistDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Information on a specific auto receptionist",
		Attributes: map[string]schema.Attribute{
			"auto_receptionist_id": schema.StringAttribute{
				MarkdownDescription: "Auto receptionist ID. The unique identifier of the auto receptionist",
				Required:            true,
			},
		},
	}
}

func (d *phoneAutoReceptionistDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.crud = NewPhoneReceptionistCrud(data.PhoneMasterClient)
}

func (d *phoneAutoReceptionistDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PhoneAutoReceptionistModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := d.crud.Read(ctx, data.AutoReceptionistID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone auto receptionist", err.Error())
		return
	}

	tflog.Info(ctx, "read phone auto receptionist", map[string]interface{}{
		"auto_receptionist_id": model.AutoReceptionistID.ValueString(),
	})

	diags := resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
