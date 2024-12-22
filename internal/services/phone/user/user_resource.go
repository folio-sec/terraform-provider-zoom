package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/folio-sec/terraform-provider-zoom/internal/schema/customvalidator"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

var (
	_ resource.Resource                = &tfResource{}
	_ resource.ResourceWithConfigure   = &tfResource{}
	_ resource.ResourceWithImportState = &tfResource{}
)

type resourceModel struct {
	UserID             types.String `tfsdk:"user_id"`
	CostCenter         types.String `tfsdk:"cost_center"`
	EmergencyAddressID types.String `tfsdk:"emergency_address_id"`
	ExtensionID        types.String `tfsdk:"extension_id"`
	ExtensionNumber    types.Int64  `tfsdk:"extension_number"`
	PhoneUserID        types.String `tfsdk:"phone_user_id"`
	SiteID             types.String `tfsdk:"site_id"`
	TemplateID         types.String `tfsdk:"template_id"`
}

func NewPhoneUserResource() resource.Resource {
	return &tfResource{}
}

type tfResource struct {
	crud        *crud
	phoneClient *zoomphone.Client
}

func (r *tfResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_user"
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
	r.crud = newCrud(data.PhoneClient, data.UserClient)
	r.phoneClient = data.PhoneClient
}

func (r *tfResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: `Manages a user within Zoom Phone.

## API Permissions

The following API permissions are required in order to use this resource.

This resource requires the ` + strings.Join([]string{
			"`user:update:user:admin`",
			"`phone:read:user:admin`",
			"`phone:update:user:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: "The ID of the Zoom user.",
			},
			"cost_center": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The cost center name.",
			},
			"emergency_address_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The emergency address ID.",
			},
			"extension_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The extension ID.",
			},
			"extension_number": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					customvalidator.Int64GreaterThan(99),
				},
				MarkdownDescription: "The extension ID. Allowed more than 3 digits. Normally, the number of digits is limited to 6, but you might be increased by contacting support.",
			},
			"phone_user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Phone user.",
			},
			"site_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The unique identifier of the [site](https://support.zoom.us/hc/en-us/articles/360020809672z) where the user should be moved or assigned.",
			},
			"template_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: "The settings template ID. If the `site_id` is set, look for the template site with the value of the `site_id`. The template ID has precedence and the policy will be ignored even if the policy field is set.",
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

	_, err := r.crud.create(ctx, createDto{
		zoomUserID: plan.UserID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone user",
			err.Error(),
		)
		return
	}

	// Wait until a phone user is created, as it is created asynchronously on Zoom.
	if _, err := util.WaitForUpdateWithTimeout(ctx, 3*time.Minute, func(ctx context.Context) (*bool, error) {
		_, err := r.phoneClient.PhoneUser(ctx, zoomphone.PhoneUserParams{
			UserId: plan.UserID.ValueString(),
		})
		if err != nil {
			if errRes, ok := lo.ErrorsAs[*zoomphone.ErrorResponseStatusCode](err); ok {
				if errRes.StatusCode == http.StatusNotFound {
					return lo.ToPtr(false), nil
				}
			}
			return nil, err
		}

		return lo.ToPtr(true), nil
	}); err != nil {
		_ = r.crud.delete(ctx, plan.UserID)
		resp.Diagnostics.AddError(
			"Waiting for a phone user to be created, but it might have been never created.",
			err.Error(),
		)
		return
	}

	if err := r.update(ctx, plan); err != nil {
		_ = r.crud.delete(ctx, plan.UserID)
		resp.Diagnostics.AddError("Error updating phone user on creating", err.Error())
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone user on creating", err.Error())
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
		resp.Diagnostics.AddError("Error reading phone user", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, plan resourceModel) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, plan.UserID)
	if err != nil {
		return nil, err
	}

	return &resourceModel{
		UserID:             plan.UserID,
		CostCenter:         dto.costCenter,
		EmergencyAddressID: dto.emergencyAddressID,
		ExtensionID:        dto.extensionID,
		ExtensionNumber:    dto.extensionNumber,
		PhoneUserID:        dto.phoneUserID,
		SiteID:             dto.siteID,
		TemplateID:         plan.TemplateID,
	}, nil
}

func (r tfResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.update(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Error updating phone user on updating", err.Error())
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone user on updating", err.Error())
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) update(ctx context.Context, plan resourceModel) error {
	if err := r.crud.update(ctx, updateDto{
		zoomUserID:         plan.UserID,
		emergencyAddressID: plan.EmergencyAddressID,
		extensionNumber:    plan.ExtensionNumber,
		siteID:             plan.SiteID,
		templateID:         plan.TemplateID,
	}); err != nil {
		return fmt.Errorf("error updating phone user: %v", err)
	}

	return nil
}

func (r *tfResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.crud.delete(ctx, state.UserID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone user",
			fmt.Sprintf(
				"Could not delete phone user %s, unexpected error: %s",
				state.UserID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted a phone user", map[string]interface{}{
		"user_id":       state.UserID.ValueString(),
		"phone_user_id": state.PhoneUserID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("user_id"), req, resp)
}
