package blockedlist

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

func NewPhoneBlockedListDataSource() datasource.DataSource {
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
	d.crud = newCrud(data.PhoneMasterClient)
}

func (d *tfDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_blocked_list"
}

func (d *tfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {

	resp.Schema = schema.Schema{
		MarkdownDescription: `A Zoom account owner or a user with the admin privilege can block phone numbers for phone users in an account.
Blocked numbers can be inbound (numbers will be blocked from calling in) and outbound (phone users in your account won't be able to dial those numbers).
Blocked callers will hear a generic message stating that the person they are calling is not available.

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:blocked_list:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier of the blocked list.",
			},
			"block_type": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `Block type.
  - inbound: The blocked number or numbers with the specifie prefix are prevented from calling in to phone users.
  - outbound: The phone users  are prevented from calling the blocked number or numbers with the specified prefix.
  - threat
`,
			},
			"comment": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Provide a comment to help you identify the blocked number or prefix. Constraints: Max 255 chars.",
			},
			"match_type": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `Indicates the match type for the blocked list. The values can be one of the following:
  - phoneNumber: Indicates that only a specific phone number that is shown in the phone_number field is blocked.
  - prefix: Indicates that all numbers starting with prefix that is shown in the phone_number field are blocked.`,
			},
			"phone_number": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The phone number or the prefix number that is blocked based on the `match_type`. Displayed in E164 format.",
			},
			"status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: `Indicates whether the blocking is active or inactive.
  - active: The blocked list is active.
  - inactive: The blocked list is inactive.`,
			},
		},
	}
}

type dataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	BlockType   types.String `tfsdk:"block_type"`
	Comment     types.String `tfsdk:"comment"`
	MatchType   types.String `tfsdk:"match_type"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	Status      types.String `tfsdk:"status"`
}

func (d *tfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dto, err := d.crud.read(ctx, data.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone blocked list", err.Error())
		return
	}

	tflog.Info(ctx, "read phone blocked list", map[string]interface{}{
		"blocked_list_id": dto.blockedListID.ValueString(),
	})

	output := dataSourceModel{
		ID:          dto.blockedListID,
		BlockType:   dto.blockType,
		Comment:     dto.comment,
		MatchType:   dto.matchType,
		PhoneNumber: dto.phoneNumber,
		Status:      dto.status,
	}
	diags := resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
