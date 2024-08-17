package callqueue

import (
	"context"
	"fmt"

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

func NewPhoneCallQueueResource() resource.Resource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_call_queue"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Call queues allow you to route incoming calls to a group of users. For instance, you can use call queues to route calls to various departments in your organization such as sales, engineering, billing, customer service etc.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Unique identifier of the Call Queue.",
			},
			"cost_center": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cost center name.",
			},
			"department": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Department name.",
			},
			"extension_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Extension ID.",
			},
			"extension_number": schema.Int64Attribute{
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				MarkdownDescription: "Extension number of the call queue.",
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(32),
				},
				MarkdownDescription: "Name of the call queue.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(32),
				},
				MarkdownDescription: "Description for the Call Queue.",
			},
			"site_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Unique identifier of the [site](https://support.zoom.us/hc/en-us/articles/360020809672-Managing-Multiple-Sites) where the Call Queue is assigned.",
			},
			"status": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("active"),
				Validators: []validator.String{
					stringvalidator.OneOf("active", "inactive"),
				},
				MarkdownDescription: `Status of the Call Queue.
  - Allowed: active┃inactive`,
			},
		},
	}
}

type resourceModel struct {
	ID              types.String `tfsdk:"id"`
	CostCenter      types.String `tfsdk:"cost_center"`
	Department      types.String `tfsdk:"department"`
	ExtensionID     types.String `tfsdk:"extension_id"`
	ExtensionNumber types.Int64  `tfsdk:"extension_number"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
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

	output, err := r.read(ctx, state.ID, state.Description)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone call queue", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, callQueueId, description types.String) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, callQueueId)
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
		ID:              dto.callQueueID,
		CostCenter:      dto.costCenter,
		Department:      dto.department,
		ExtensionID:     dto.extensionID,
		ExtensionNumber: dto.extensionNumber,
		Name:            dto.name,
		// Description: dto.description, // get api not supported yet
		Description: description,
		SiteID:      siteID,
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
		name:            plan.Name,
		siteID:          plan.SiteID,
		costCenter:      plan.CostCenter,
		department:      plan.Department,
		extensionNumber: plan.ExtensionNumber,
		description:     plan.Description,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call queue",
			err.Error(),
		)
		return
	}

	// some fields cannot set on create api, so update them
	if err := r.crud.update(ctx, &updateDto{
		callQueueID:     ret.callQueueID,
		siteID:          plan.SiteID,
		costCenter:      plan.CostCenter,
		department:      plan.Department,
		extensionNumber: plan.ExtensionNumber,
		name:            plan.Name,
		description:     plan.Description,
		status:          plan.Status,
	}); err != nil {
		// TODO change delete logic with marking resource as taint
		_ = r.crud.delete(ctx, ret.callQueueID)
		resp.Diagnostics.AddError(
			"Error creating phone call queue on updating",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, ret.callQueueID, plan.Description)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone call queue on reading", err.Error())
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
			"Error updating phone call queue",
			"Error updating phone call queue",
		)
		return
	}

	if err := r.crud.update(ctx, &updateDto{
		callQueueID:     plan.ID,
		siteID:          plan.SiteID,
		costCenter:      plan.CostCenter,
		department:      plan.Department,
		extensionNumber: plan.ExtensionNumber,
		name:            plan.Name,
		description:     plan.Description,
		status:          plan.Status,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone call queue",
			fmt.Sprintf(
				"Could not update phone call queue %s, unexpected error: %s",
				plan.ID.ValueString(),
				err,
			),
		)
		return
	}

	output, err := r.read(ctx, plan.ID, plan.Description)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone call queue", err.Error())
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
			"Error deleting phone call queue",
			fmt.Sprintf(
				"Could not delete phone call queue %s, unexpected error: %s",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone call queue", map[string]interface{}{
		"call_queue_id": state.ID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
