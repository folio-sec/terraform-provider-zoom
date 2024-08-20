package callqueuephonenumbers

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

func NewPhoneCallQueuePhoneNumbersResource() resource.Resource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_call_queue_phone_numbers"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `After [buying phone number(s)](https://support.zoom.us/hc/en-us/articles/360020808292#h_007ec8c2-0914-4265-8351-96ab23efa3ad), you can assign it, allowing callers to directly dial a number to reach a [call queue](https://support.zoom.us/hc/en-us/articles/360021524831-Managing-Call-Queues).

## API Permissions
The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:call_queue:admin`",
			"`phone:read:list_call_queues:admin`",
			"`phone:read:list_numbers:admin`",
			"`phone:write:call_queue_number:admin`",
			"`phone:delete:call_queue_number:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"call_queue_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique identifier of the Call Queue.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"phone_numbers": schema.SetNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Unique identifier of the number. Provide either the `id` or the `number` field. ",
						},
						"number": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Phone number e.g. `+12058945456`. Provide either the `id` or the `number` field. ",
						},
						"source": schema.StringAttribute{
							Optional: true,
							Computed: true,
							MarkdownDescription: `Source
  - Allowed: internal┃external`,
							Validators: []validator.String{
								stringvalidator.OneOf("internal", "external"),
							},
						},
					},
				},
			},
		},
	}
}

type resourceModel struct {
	CallQueueID  types.String                `tfsdk:"call_queue_id"`
	PhoneNumbers []*resourceModelPhoneNumber `tfsdk:"phone_numbers"`
}

type resourceModelPhoneNumber struct {
	ID     types.String `tfsdk:"id"`
	Number types.String `tfsdk:"number"`
	Source types.String `tfsdk:"source"`
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
		resp.Diagnostics.AddError("Error reading phone call queue phone numbers", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) read(ctx context.Context, plan resourceModel) (*resourceModel, error) {
	dto, err := r.crud.read(ctx, plan.CallQueueID)
	if err != nil {
		return nil, err
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	phoneNumbers := lo.Map(dto.phoneNumbers, func(p *readDtoPhoneNumber, _index int) *resourceModelPhoneNumber {
		return &resourceModelPhoneNumber{
			ID:     p.id,
			Number: p.number,
			Source: p.source,
		}
	})
	return &resourceModel{
		CallQueueID:  plan.CallQueueID,
		PhoneNumbers: phoneNumbers,
	}, nil
}

func (r *tfResource) sync(ctx context.Context, plan resourceModel) error {
	asis, err := r.read(ctx, plan)
	if err != nil {
		return err
	}
	if asis == nil {
		return fmt.Errorf("calal queue not found %s", plan.CallQueueID.ValueString())
	}

	// 0. plan validation (it might be better to move into validator)
	for _, p := range plan.PhoneNumbers {
		if p.ID.ValueString() == "" && p.Number.ValueString() == "" {
			return fmt.Errorf("either `id` or `number` must be specified on phone number")
		}
	}

	// 1. unassign phone numbers = asis - plan
	var unassignPhoneNumberIDs []types.String
	for _, asisPhoneNumber := range asis.PhoneNumbers {
		planExisted := lo.ContainsBy(plan.PhoneNumbers, func(planItem *resourceModelPhoneNumber) bool {
			// allow either id or number parameter
			return planItem.ID == asisPhoneNumber.ID || planItem.Number == asisPhoneNumber.Number
		})
		if !planExisted {
			unassignPhoneNumberIDs = append(unassignPhoneNumberIDs, asisPhoneNumber.ID)
		}
	}
	if err = r.crud.unassign(ctx, &unassignDto{
		callQueueID:    plan.CallQueueID,
		phoneNumberIDs: unassignPhoneNumberIDs,
	}); err != nil {
		return err
	}

	// 2. assign phone numbers = plan - asis
	var assignPhoneNumberIDs []types.String
	var assignPhoneNumbers []types.String
	for _, planPhoneNumber := range plan.PhoneNumbers {
		asisExisted := lo.ContainsBy(asis.PhoneNumbers, func(asisItem *resourceModelPhoneNumber) bool {
			// allow either id or number parameter
			return asisItem.ID == planPhoneNumber.ID || asisItem.Number == planPhoneNumber.Number
		})
		if !asisExisted {
			if planPhoneNumber.ID.ValueString() != "" {
				assignPhoneNumberIDs = append(assignPhoneNumberIDs, planPhoneNumber.ID)
			} else {
				assignPhoneNumbers = append(assignPhoneNumbers, planPhoneNumber.Number)
			}
		}
	}
	if err = r.crud.assign(ctx, &assignDto{
		callQueueID:    plan.CallQueueID,
		phoneNumberIDs: assignPhoneNumberIDs,
		phoneNumbers:   assignPhoneNumbers,
	}); err != nil {
		return err
	}
	return nil
}

func (r *tfResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.sync(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call queue phone numbers",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone call queue phone numbers on reading", err.Error())
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
			"Error updating phone call queue phone numbers on get plan",
			"Error updating phone call queue phone numbers",
		)
		return
	}

	if err := r.sync(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone call queue phone numbers",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone call queue phone numbers", err.Error())
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

	if err := r.crud.unassignAll(ctx, state.CallQueueID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone call queue phone numbers",
			fmt.Sprintf(
				"Could not delete phone call queue phone numbers %s, unexpected error: %s",
				state.CallQueueID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone call queue phone numbers", map[string]interface{}{
		"call_queue_id": state.CallQueueID.ValueString(),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("call_queue_id"), req, resp)
}
