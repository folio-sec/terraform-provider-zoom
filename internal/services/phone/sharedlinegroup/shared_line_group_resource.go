package sharedlinegroupgroup

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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

func NewPhoneSharedLineGroupResource() resource.Resource {
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
	r.crud = newCrud(data.PhoneMasterClient)
}

func (r *tfResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_shared_line_group"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `A [shared line group](https://support.zoom.us/hc/en-us/articles/360038850792) allows Zoom Phone admins to share a phone number and extension with a group of phone users or common areas. This gives members of the shared line group access to the group's direct phone number and voicemail.

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:shared_line_group:admin`",
			"`phone:write:shared_line_group:admin`",
			"`phone:update:shared_line_group:admin`",
			"`phone:delete:shared_line_group:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The unique identifier of the shared line group.",
			},
			"display_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(200),
				},
				MarkdownDescription: "The name to identify the shared line group.",
			},
			"extension_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Extension ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"extension_number": schema.Int64Attribute{
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				MarkdownDescription: "Extension number of the shared line group.",
			},
			"primary_number": schema.StringAttribute{
				// primary_number is managed by "shared_line_group_phone_numbers" resource
				// so "shared_line_group.primary_number" should be read only value.
				Computed: true,
				MarkdownDescription: `If you have multiple direct phone numbers assigned to the shared line group, this is the primary number selected for desk phones.
The primary number shares the same line as the extension number. This means if a caller is routed to the shared line group through an auto receptionist, the line associated with the primary number will be used.`,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"site_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Unique identifier of the [site](https://support.zoom.us/hc/en-us/articles/360020809672-Managing-Multiple-Sites) where the shared line group is assigned.",
				// update api doesn't support site_id, so replace on updating
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("active"),
				Validators: []validator.String{
					stringvalidator.OneOf("active", "inactive"),
				},
				MarkdownDescription: `The status of the shared line group.
  - Allowed: active┃inactive`,
			},
		},
	}
}

type resourceModel struct {
	ID              types.String `tfsdk:"id"`
	DisplayName     types.String `tfsdk:"display_name"`
	ExtensionID     types.String `tfsdk:"extension_id"`
	ExtensionNumber types.Int64  `tfsdk:"extension_number"`
	PrimaryNumber   types.String `tfsdk:"primary_number"`
	SiteID          types.String `tfsdk:"site_id"`
	Status          types.String `tfsdk:"status"`
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
		resp.Diagnostics.AddError("Error reading phone shared line group", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, sharedLineGroupId types.String) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, sharedLineGroupId)
	if err != nil {
		return nil, fmt.Errorf("error read: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	siteID := types.StringNull()
	if dto.site != nil {
		siteID = dto.site.id
	}
	return &resourceModel{
		ID:              dto.sharedLineGroupID,
		DisplayName:     dto.displayName,
		ExtensionID:     dto.extensionID,
		ExtensionNumber: dto.extensionNumber,
		PrimaryNumber:   dto.primaryNumber,
		Status:          dto.status,
		SiteID:          siteID,
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
		displayName:     plan.DisplayName,
		extensionNumber: plan.ExtensionNumber,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone shared line group",
			err.Error(),
		)
		return
	}

	// some fields cannot set on create api, so update them
	if err := r.crud.update(ctx, &updateDto{
		sharedLineGroupID: ret.sharedLineGroupID,
		extensionNumber:   plan.ExtensionNumber,
		displayName:       plan.DisplayName,
		status:            plan.Status,
	}); err != nil {
		// TODO change delete logic with marking resource as taint
		_ = r.crud.delete(ctx, ret.sharedLineGroupID)
		resp.Diagnostics.AddError(
			"Error creating phone shared line group on updating",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, ret.sharedLineGroupID)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone shared line group on reading", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error updating phone shared line group",
			"Error updating phone shared line group",
		)
		return
	}

	if err := r.crud.update(ctx, &updateDto{
		sharedLineGroupID: plan.ID,
		extensionNumber:   plan.ExtensionNumber,
		displayName:       plan.DisplayName,
		status:            plan.Status,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone shared line group",
			fmt.Sprintf(
				"Could not update phone shared line group %s, unexpected error: %s",
				plan.ID.ValueString(),
				err,
			),
		)
		return
	}

	output, err := r.read(ctx, plan.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone shared line group", err.Error())
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
			"Error deleting phone shared line group",
			fmt.Sprintf(
				"Could not delete phone shared line group %s, unexpected error: %s",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone shared line group", map[string]interface{}{
		"shared_line_group_id": state.ID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
