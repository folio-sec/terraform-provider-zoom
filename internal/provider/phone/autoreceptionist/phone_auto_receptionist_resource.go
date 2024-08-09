package autoreceptionist

import (
	"context"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &phoneAutoReceptionistResource{}
	_ resource.ResourceWithConfigure   = &phoneAutoReceptionistResource{}
	_ resource.ResourceWithImportState = &phoneAutoReceptionistResource{}
)

func NewPhoneReceptionistResource() resource.Resource {
	return &phoneAutoReceptionistResource{}
}

type phoneAutoReceptionistResource struct {
	crud *PhoneAutoReceptionistCrud
}

func (r *phoneAutoReceptionistResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_auto_receptionist"
}

// Configure adds the provider configured client to the resource.
func (r *phoneAutoReceptionistResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.crud = NewPhoneReceptionistCrud(data.PhoneMasterClient)
}

// Create creates the resource and sets the initial Terraform state.
func (r *phoneAutoReceptionistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PhoneAutoReceptionistModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ret, err := r.crud.Create(ctx, PhoneAutoReceptionistCreateDto{
		Name: data.Name,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone auto receptionist",
			err.Error(),
		)
		return
	}
	err = r.crud.Update(ctx, PhoneAutoReceptionistUpdateDto{
		AutoReceptionistID:  ret.ID,
		CostCenter:          data.CostCenter,
		Department:          data.Department,
		ExtensionNumber:     ret.ExtensionNumber,
		Name:                ret.Name,
		Timezone:            data.Timezone,
		AudioPromptLanguage: data.AudioPromptLanguage,
	})
	if err != nil {
		// TODO mark resource as taint
		resp.Diagnostics.AddError(
			"Error creating phone auto receptionist on updating",
			err.Error(),
		)
		return
	}

	model := &PhoneAutoReceptionistModel{
		AutoReceptionistID:  ret.ID,
		CostCenter:          data.CostCenter,
		Department:          data.Department,
		ExtensionNumber:     ret.ExtensionNumber,
		Name:                ret.Name,
		Timezone:            data.Timezone,
		AudioPromptLanguage: data.AudioPromptLanguage,
	}
	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *phoneAutoReceptionistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PhoneAutoReceptionistModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.crud.Read(ctx, data.AutoReceptionistID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone auto receptionist", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *phoneAutoReceptionistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PhoneAutoReceptionistModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error getting phone auto receptionist",
			"Error getting phone auto receptionist",
		)
		return
	}

	var state PhoneAutoReceptionistModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *phoneAutoReceptionistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PhoneAutoReceptionistModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.crud.Delete(ctx, data.AutoReceptionistID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone auto receptionist",
			fmt.Sprintf(
				"Could not delete phone auto receptionist %s, unexpected error: %s",
				data.AutoReceptionistID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone auto receptionist", map[string]interface{}{
		"auto_receptionist_id": data.AutoReceptionistID.ValueString(),
	})
}

func (r *phoneAutoReceptionistResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Information on a specific auto receptionist",
		Attributes: map[string]schema.Attribute{
			"auto_receptionist_id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cost_center": schema.StringAttribute{
				Optional: true,
			},
			"department": schema.StringAttribute{
				Optional: true,
			},
			"extension_number": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"timezone": schema.StringAttribute{
				Optional: true,
			},
			"audio_prompt_language": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *phoneAutoReceptionistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model, err := r.crud.Read(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading phone auto receptionist",
			fmt.Sprintf("Could not get phone auto receptionist %s, unexpected error: %s",
				req.ID,
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "imported phone auto receptionist", map[string]interface{}{
		"auro_receptionist_id": req.ID,
	})

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
