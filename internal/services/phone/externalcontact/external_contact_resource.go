package externalcontact

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &tfResource{}
	_ resource.ResourceWithConfigure   = &tfResource{}
	_ resource.ResourceWithImportState = &tfResource{}
)

func NewPhoneExternalContactResource() resource.Resource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_external_contact"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `External contact's information.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:external_contact:admin`",
			"`phone:write:external_contact:admin`",
			"`phone:update:external_contact:admin`",
			"`phone:delete:external_contact:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The customer-configured external contact ID. It is recommended that you use a primary key from the original phone system. If you do not use this parameter, the API automatically generates an `external_contact_id`.",
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
				MarkdownDescription: "The external contact's username or extension display name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The external contact's description.",
			},
			"extension_number": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The external contact's extension number.",
			},
			"email": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
				MarkdownDescription: "The external contact's email address.",
			},
			"phone_numbers": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				MarkdownDescription: "The external contact's phone numbers. This value must be in [E.164](https://en.wikipedia.org/wiki/E.164) format. If you do not provide an extension number, you must provide this value.",
			},
			"auto_call_recorded": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to allow the automatic call recording.",
			},
			"external_contact_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Zoom-generated external contact ID.",
			},
			"routing_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The external contact's SIP group, to define the call routing path. This is for customers that use SIP trunking.",
			},
		},
	}
}

type resourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	Description       types.String   `tfsdk:"description"`
	ExtensionNumber   types.String   `tfsdk:"extension_number"`
	Email             types.String   `tfsdk:"email"`
	PhoneNumbers      []types.String `tfsdk:"phone_numbers"`
	AutoCallRecorded  types.Bool     `tfsdk:"auto_call_recorded"`
	ExternalContactID types.String   `tfsdk:"external_contact_id"`
	RoutingPath       types.String   `tfsdk:"routing_path"`
}

func (r *tfResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := r.read(ctx, state.ExternalContactID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading phone external contact", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, externalContactID types.String) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, externalContactID)
	if err != nil {
		return nil, fmt.Errorf("error read: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	return &resourceModel{
		ID:                dto.id,
		Name:              dto.name,
		Description:       dto.description,
		ExtensionNumber:   dto.extensionNumber,
		Email:             dto.email,
		PhoneNumbers:      dto.phoneNumbers,
		AutoCallRecorded:  dto.autoCallRecorded,
		ExternalContactID: dto.externalContactID,
		RoutingPath:       dto.routingPath,
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
		description:       plan.Description,
		email:             plan.Email,
		extensionNumber:   plan.ExtensionNumber,
		externalContactID: plan.ExternalContactID,
		id:                plan.ID,
		name:              plan.Name,
		phoneNumbers:      plan.PhoneNumbers,
		routingPath:       plan.RoutingPath,
		autoCallRecorded:  plan.AutoCallRecorded,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone external contact",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, ret.externalContactID)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone external contact on reading", err.Error())
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
			"Error updating phone external contact",
			"Error updating phone external contact",
		)
		return
	}

	if err := r.crud.update(ctx, &updateDto{
		externalContactID: plan.ExternalContactID,
		description:       plan.Description,
		email:             plan.Email,
		extensionNumber:   plan.ExtensionNumber,
		id:                plan.ID,
		name:              plan.Name,
		phoneNumbers:      plan.PhoneNumbers,
		routingPath:       plan.RoutingPath,
		autoCallRecorded:  plan.AutoCallRecorded,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone external contact",
			fmt.Sprintf(
				"Could not update phone external contact %s, unexpected error: %s",
				plan.ID.ValueString(),
				err,
			),
		)
		return
	}

	output, err := r.read(ctx, plan.ExternalContactID)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone external contact", err.Error())
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

	if err := r.crud.delete(ctx, state.ExternalContactID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone external contact",
			fmt.Sprintf(
				"Could not delete phone external contact %s, unexpected error: %s",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone external contact", map[string]interface{}{
		"external_contact_id": state.ExternalContactID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("external_contact_id"), req, resp)
}
