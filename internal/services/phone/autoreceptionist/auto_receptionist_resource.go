package autoreceptionist

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &tfResource{}
	_ resource.ResourceWithConfigure   = &tfResource{}
	_ resource.ResourceWithImportState = &tfResource{}
)

func NewPhoneAutoReceptionistResource() resource.Resource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_auto_receptionist"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Auto receptionists answer calls with a personalized recording and routes calls to a phone user, call queue, common area, voicemail or an IVR system.

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:auto_receptionist:admin`",
			"`phone:write:auto_receptionist:admin`",
			"`phone:update:auto_receptionist:admin`",
			"`phone:delete:auto_receptionist:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Auto receptionist ID. The unique identifier of the auto receptionist.",
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
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				MarkdownDescription: "Extension number of the auto receptionist.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the auto receptionist.",
			},
			"timezone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "[Timezone](https://marketplace.zoom.us/docs/api-reference/other-references/abbreviation-lists#timezones) of the Auto Receptionist.",
			},
			"audio_prompt_language": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: `The language for all default audio prompts for the auto receptionist.
  - Allowed: en-US┃en-GB┃es-US┃fr-CA┃da-DK┃de-DE┃es-ES┃fr-FR┃it-IT┃nl-NL┃pt-PT┃ja┃ko-KR┃pt-BR┃zh-CN
`,
			},
		},
	}
}

type resourceModel struct {
	ID                  types.String `tfsdk:"id"`
	CostCenter          types.String `tfsdk:"cost_center"`
	Department          types.String `tfsdk:"department"`
	ExtensionID         types.String `tfsdk:"extension_id"`
	ExtensionNumber     types.Int64  `tfsdk:"extension_number"`
	Name                types.String `tfsdk:"name"`
	Timezone            types.String `tfsdk:"timezone"`
	AudioPromptLanguage types.String `tfsdk:"audio_prompt_language"`
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
		resp.Diagnostics.AddError("Error reading phone auto receptionist", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, autoReceptionistId types.String) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, autoReceptionistId)
	if err != nil {
		return nil, fmt.Errorf("error read: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	return &resourceModel{
		ID:                  dto.autoReceptionistID,
		CostCenter:          dto.costCenter,
		Department:          dto.department,
		ExtensionID:         dto.extensionID,
		ExtensionNumber:     dto.extensionNumber,
		Name:                dto.name,
		Timezone:            dto.timezone,
		AudioPromptLanguage: dto.audioPromptLanguage,
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
		name: plan.Name,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone auto receptionist",
			err.Error(),
		)
		return
	}
	err = r.crud.update(ctx, &updateDto{
		autoReceptionistID:  ret.autoReceptionistID,
		costCenter:          plan.CostCenter,
		department:          plan.Department,
		extensionNumber:     plan.ExtensionNumber,
		name:                plan.Name,
		timezone:            plan.Timezone,
		audioPromptLanguage: plan.AudioPromptLanguage,
	})
	if err != nil {
		// TODO change delete logic with marking resource as taint
		_ = r.crud.delete(ctx, ret.autoReceptionistID)
		resp.Diagnostics.AddError(
			"Error creating phone auto receptionist on updating",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, ret.autoReceptionistID)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone auto receptionist on reading", err.Error())
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
			"Error updating phone auto receptionist",
			"Error updating phone auto receptionist",
		)
		return
	}

	if err := r.crud.update(ctx, &updateDto{
		autoReceptionistID:  plan.ID,
		costCenter:          plan.CostCenter,
		department:          plan.Department,
		extensionNumber:     plan.ExtensionNumber,
		name:                plan.Name,
		timezone:            plan.Timezone,
		audioPromptLanguage: plan.AudioPromptLanguage,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone auto receptionist",
			fmt.Sprintf(
				"Could not update phone auto receptionist %s, unexpected error: %s",
				plan.ID.ValueString(),
				err,
			),
		)
		return
	}

	output, err := r.read(ctx, plan.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone auto receptionist", err.Error())
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
			"Error deleting phone auto receptionist",
			fmt.Sprintf(
				"Could not delete phone auto receptionist %s, unexpected error: %s",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone auto receptionist", map[string]interface{}{
		"auto_receptionist_id": state.ID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
