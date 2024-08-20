package blockedlist

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &tfResource{}
	_ resource.ResourceWithConfigure   = &tfResource{}
	_ resource.ResourceWithImportState = &tfResource{}
)

func NewPhoneBlockedListResource() resource.Resource {
	return &tfResource{}
}

type tfResource struct {
	crud *crud
}

func (r *tfResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.crud = newCrud(data.PhoneClient)
}

func (r *tfResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_blocked_list"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `A Zoom account owner or a user with the admin privilege can block phone numbers for phone users in an account.
Blocked numbers can be inbound (numbers will be blocked from calling in) and outbound (phone users in your account won't be able to dial those numbers).
Blocked callers will hear a generic message stating that the person they are calling is not available.

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:blocked_list:admin`",
			"`phone:write:blocked_list:admin`",
			// "`phone:update:blocked_list:admin`", // PATCH api hasn't be provided yet on openapi spec
			"`phone:delete:blocked_list:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the blocked list.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"block_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `Block type.
  - inbound: The blocked number or numbers with the specifie prefix are prevented from calling in to phone users.
  - outbound: The phone users  are prevented from calling the blocked number or numbers with the specified prefix.
`,
				// PATCH blocked_list hasn't be provided yet, so we just do delete/create on update.
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.OneOf("inbound", "outbound", "threat"),
				},
			},
			"comment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Provide a comment to help you identify the blocked number or prefix. Constraints: Max 255 chars.",
				// PATCH blocked_list hasn't be provided yet, so we just do delete/create on update.
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"match_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `Indicates the match type for the blocked list. The values can be one of the following:
  - phoneNumber: Indicates that only a specific phone number that is shown in the phone_number field is blocked.
  - prefix: Indicates that all numbers starting with prefix that is shown in the phone_number field are blocked.`,
				// PATCH blocked_list hasn't be provided yet, so we just do delete/create on update.
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.OneOf("phoneNumber", "prefix"),
				},
			},
			"phone_number": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The phone number or the prefix number that is blocked based on the `match_type`. Displayed in E164 format.",
				// PATCH blocked_list hasn't be provided yet, so we just do delete/create on update.
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: `Indicates whether the blocking is active or inactive.
  - active: The blocked list is active.
  - inactive: The blocked list is inactive.`,
				Default: stringdefault.StaticString("active"),
				// PATCH blocked_list hasn't be provided yet, so we just do delete/create on update.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("active", "inactive"),
				},
			},
		},
	}
}

type resourceModel struct {
	ID          types.String `tfsdk:"id"`
	BlockType   types.String `tfsdk:"block_type"`
	Comment     types.String `tfsdk:"comment"`
	MatchType   types.String `tfsdk:"match_type"`
	PhoneNumber types.String `tfsdk:"phone_number"`
	Status      types.String `tfsdk:"status"`
}

func (r *tfResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.read(ctx, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone blocked list", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, blockedListId types.String) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, blockedListId)
	if err != nil {
		return nil, fmt.Errorf("error read: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	return &resourceModel{
		ID:          dto.blockedListID,
		BlockType:   dto.blockType,
		Comment:     dto.comment,
		MatchType:   dto.matchType,
		PhoneNumber: dto.phoneNumber,
		Status:      dto.status,
	}, nil
}

func (r *tfResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ret, err := r.crud.create(ctx, &createDto{
		blockType:   plan.BlockType,
		comment:     plan.Comment,
		matchType:   plan.MatchType,
		phoneNumber: plan.PhoneNumber,
		status:      plan.Status,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone blocked list",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, ret.blockedListID)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone blocked list on reading", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// PATCH blocked_list hasn't be provided yet, so we just do delete/create on update.
	resp.Diagnostics.AddError(
		"Error updating phone blocked list",
		"blocked list update is not supported, please delete and recreate the resource, this caused by provider issue",
	)
	return
}

func (r *tfResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.crud.delete(ctx, state.ID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone blocked list",
			fmt.Sprintf(
				"Could not delete phone blocked list %s, unexpected error: %s",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone blocked list", map[string]interface{}{
		"auto_receptionist_id": state.ID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
