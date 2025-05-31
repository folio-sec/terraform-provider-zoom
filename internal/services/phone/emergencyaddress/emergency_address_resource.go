package emergencyaddress

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &tfResource{}
	_ resource.ResourceWithImportState = &tfResource{}
	_ resource.ResourceWithModifyPlan  = &tfResource{}
)

func NewEmergencyAddressResource() resource.Resource {
	return &tfResource{}
}

type tfResource struct {
	crud *crud
}

type resourceModel struct {
	ID           types.String `tfsdk:"id"`
	AddressLine1 types.String `tfsdk:"address_line1"`
	AddressLine2 types.String `tfsdk:"address_line2"`
	City         types.String `tfsdk:"city"`
	Country      types.String `tfsdk:"country"`
	IsDefault    types.Bool   `tfsdk:"is_default"`
	SiteID       types.String `tfsdk:"site_id"`
	StateCode    types.String `tfsdk:"state_code"`
	Zip          types.String `tfsdk:"zip"`
	Status       types.Int32  `tfsdk:"status"`
	Level        types.Int32  `tfsdk:"level"`
	UserID       types.String `tfsdk:"user_id"`
}

func (r *tfResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_emergency_address"
}

func (r *tfResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*shared.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected ProviderData Source Configure Type",
			fmt.Sprintf("Expected *provider.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.crud = newCrud(providerData.PhoneClient)
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownSeparatorForList := "\n  "

	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a emergency address within Zoom Phone.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:emergency_address:admin`",
			"`phone:write:emergency_address:admin`",
			"`phone:update:emergency_address:admin`",
			"`phone:delete:emergency_address:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The emergency address ID.",
			},
			"address_line1": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The emergency address line 1.",
			},
			"address_line2": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The emergency address line 2.",
			},
			"city": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The emergency address city.",
			},
			"country": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
					stringvalidator.LengthAtMost(2),
				},
				MarkdownDescription: "The two-lettered country code (Alpha-2 code in ISO-3166 format) of the emergency address.",
			},
			"is_default": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Indicates whether the emergency address is default or not.",
			},
			"site_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("user_id"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "The unique identifier of the site to which this emergency address belongs. Exactly one of `site_id` or `user_id` must be specified.",
			},
			"state_code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The emergency address state code.",
			},
			"zip": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The emergency address zip code.",
			},
			"status": schema.Int32Attribute{
				Computed: true,
				Validators: []validator.Int32{
					int32validator.Between(1, 6),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
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
			"level": schema.Int32Attribute{
				Computed: true,
				Validators: []validator.Int32{
					int32validator.Between(0, 2),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The emergency address owner level." + strings.Join([]string{
					"",
					"- `0`: Account/Company-level emergency address.",
					"- `1`: User/Personal-level emergency address.",
					"- `2`: Unknown company/pending emergency address.",
				}, markdownSeparatorForList),
			},
			"user_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("site_id"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "User ID to which the personal emergency address belongs. Exactly one of `site_id` or `user_id` must be specified.",
			},
		},
	}
}

func (r *tfResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ret, err := r.crud.create(ctx, &createDto{
		addressLine1: plan.AddressLine1,
		addressLine2: plan.AddressLine2,
		city:         plan.City,
		country:      plan.Country,
		isDefault:    plan.IsDefault,
		siteID:       plan.SiteID,
		state:        plan.StateCode,
		zip:          plan.Zip,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating emergency address",
			err.Error(),
		)
		return
	}

	plan.ID = ret.id
	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating emergency address on reading", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.read(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone emergency address", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, plan resourceModel) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, plan.ID)
	if err != nil {
		return nil, fmt.Errorf("error read phone emergency address: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	return &resourceModel{
		ID:           dto.id,
		AddressLine1: dto.addressLine1,
		AddressLine2: dto.addressLine2,
		City:         dto.city,
		Country:      dto.country,
		IsDefault:    dto.isDefault,
		SiteID:       dto.site.ID,
		StateCode:    dto.stateCode,
		Zip:          dto.zip,
		Status:       dto.status,
		Level:        dto.level,
		UserID:       dto.owner.ID,
	}, nil
}

func (r *tfResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error updating phone emergency address on get plan",
			"Error updating phone emergency address",
		)
		return
	}

	if err := r.crud.update(ctx, &updateDto{
		id:           plan.ID,
		addressLine1: plan.AddressLine1,
		addressLine2: plan.AddressLine2,
		city:         plan.City,
		country:      plan.Country,
		isDefault:    plan.IsDefault,
		siteID:       plan.SiteID,
		state:        plan.StateCode,
		zip:          plan.Zip,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone emergency address",
			fmt.Sprintf(
				"Could not update phone emergency address %s, unexpected error: %s",
				plan.ID.ValueString(),
				err.Error(),
			),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone emergency address on reading", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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
			"Error deleting phone emergency address",
			fmt.Sprintf(
				"Could not delete phone emergency address %s, unexpected error: %s",
				state.ID.ValueString(),
				err.Error(),
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone emergency address", map[string]interface{}{
		"emergency_address_id": state.ID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *tfResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, state resourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if this is a delete operation (plan is null) and the resource is default and linked to a site
	if req.Plan.Raw.IsNull() && !state.SiteID.IsNull() && !state.IsDefault.IsNull() && state.IsDefault.ValueBool() {
		existing, err := r.crud.read(ctx, state.ID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading phone emergency address",
				fmt.Sprintf(
					"Could not read phone emergency address %s, unexpected error: %s",
					state.ID.ValueString(),
					err.Error(),
				),
			)
			return
		}

		// If the resource is already deleted, allow the operation
		if existing == nil {
			return
		}

		resp.Diagnostics.AddError(
			"Cannot delete linked to a site and default emergency address",
			"The emergency address is set as default and cannot be deleted. If this emergency address is linked to a site, it will be automatically deleted when the site itself is deleted.",
		)
		return
	}
}
