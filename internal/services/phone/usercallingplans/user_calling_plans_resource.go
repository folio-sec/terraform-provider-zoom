package usercallingplans

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ resource.Resource                = &tfResource{}
	_ resource.ResourceWithConfigure   = &tfResource{}
	_ resource.ResourceWithImportState = &tfResource{}
)

// See also: https://developers.zoom.us/docs/api/rest/other-references/calling-plans/
var callingPlanMapping = map[string]int32{
	"NO_FEATURE_PACKAGE":                        1,
	"INTERNATIONAL_TOLL_NUMBER":                 3,
	"INTERNATIONAL_TOLL_FREE_NUMBER":            4,
	"BYOC_NUMBER":                               5,
	"BETA_NUMBER":                               6,
	"METERED_PLAN_US_CA":                        100,
	"METERED_PLAN_AU_NZ":                        101,
	"METERED_PLAN_GB_IE":                        102,
	"METERED_EURA":                              103,
	"METERED_EURB":                              104,
	"METERED_JP":                                107,
	"UNLIMITED_PLAN_US_CA":                      200,
	"UNLIMITED_PLAN_AU_NZ":                      201,
	"UNLIMITED_PLAN_GB_IE":                      202,
	"UNLIMITED_EURA":                            203,
	"UNLIMITED_EURB":                            204,
	"UNLIMITED_JP":                              207,
	"US_CA_NUMBER":                              300,
	"AU_NZ_NUMBER":                              301,
	"GB_IE_NUMBER":                              302,
	"EURA_NUMBER":                               303,
	"EURB_NUMBER":                               304,
	"JP_NUMBER":                                 307,
	"US_CA_TOLLFREE_NUMBER":                     400,
	"AU_TOLLFREE_NUMBER":                        401,
	"GB_IE_TOLLFREE_NUMBER":                     402,
	"NZ_TOLLFREE_NUMBER":                        403,
	"GLOBAL_TOLLFREE_NUMBER":                    404,
	"BETA":                                      600,
	"UNLIMITED_DOMESTIC_SELECT":                 1000,
	"METERED_GLOBAL_SELECT":                     1001,
	"UNLIMITED_DOMESTIC_SELECT_NUMBER":          2000,
	"ZP_PRO":                                    3000,
	"BASIC":                                     3010,
	"ZP_COMMON_AREA":                            3040,
	"RESERVED_PLAN":                             3098,
	"BASIC_MIGRATED":                            3099,
	"INTERNATIONAL_SELECT_ADDON":                4000,
	"ZP_PREMIUM_ADDON":                          4010,
	"PREMIUM_NUMBER":                            5000,
	"METERED_US_CA_NUMBER_INCLUDED":             30000,
	"METERED_AU_NZ_NUMBER_INCLUDED":             30001,
	"METERED_GB_IE_NUMBER_INCLUDED":             30002,
	"METERED_EURA_NUMBER_INCLUDED":              30003,
	"METERED_EURB_NUMBER_INCLUDED":              30004,
	"METERED_JP_NUMBER_INCLUDED":                30007,
	"UNLIMITED_US_CA_NUMBER_INCLUDED":           31000,
	"UNLIMITED_AU_NZ_NUMBER_INCLUDED":           31001,
	"UNLIMITED_GB_IE_NUMBER_INCLUDED":           31002,
	"UNLIMITED_EURA_NUMBER_INCLUDED":            31003,
	"UNLIMITED_EURB_NUMBER_INCLUDED":            31004,
	"UNLIMITED_JP_NUMBER_INCLUDED":              31007,
	"UNLIMITED_DOMESTIC_SELECT_NUMBER_INCLUDED": 31005,
	"METERED_GLOBAL_SELECT_NUMBER_INCLUDED":     31006,
	"MEETINGS_PRO_UNLIMITED_US_CA":              40200,
	"MEETINGS_PRO_UNLIMITED_AU_NZ":              40201,
	"MEETINGS_PRO_UNLIMITED_GB_IE":              40202,
	"MEETINGS_PRO_UNLIMITED_JP":                 40207,
	"MEETINGS_PRO_GLOBAL_SELECT":                41000,
	"MEETINGS_PRO_PN_PRO":                       43000,
	"MEETINGS_BUS_UNLIMITED_US_CA":              50200,
	"MEETINGS_BUS_UNLIMITED_AU_NZ":              50201,
	"MEETINGS_BUS_UNLIMITED_GB_IE":              50202,
	"MEETINGS_BUS_UNLIMITED_JP":                 50207,
	"MEETINGS_BUS_GLOBAL_SELECT":                51000,
	"MEETINGS_BUS_PN_PRO":                       53000,
	"MEETINGS_ENT_UNLIMITED_US_CA":              60200,
	"MEETINGS_ENT_UNLIMITED_AU_NZ":              60201,
	"MEETINGS_ENT_UNLIMITED_GB_IE":              60202,
	"MEETINGS_ENT_UNLIMITED_JP":                 60207,
	"MEETINGS_ENT_GLOBAL_SELECT":                61000,
	"MEETINGS_ENT_PN_PRO":                       63000,
	"MEETINGS_US_CA_NUMBER_INCLUDED":            70200,
	"MEETINGS_AU_NZ_NUMBER_INCLUDED":            70201,
	"MEETINGS_GB_IE_NUMBER_INCLUDED":            70202,
	"MEETINGS_JP_NUMBER_INCLUDED":               70207,
	"MEETINGS_GLOBAL_SELECT_NUMBER_INCLUDED":    71000,
}

func NewPhoneUserCallingPlansResource() resource.Resource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_user_calling_plans"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	markdownSeparatorForList := "\n  "

	resp.Schema = schema.Schema{
		MarkdownDescription: `Assigns calling plans to a Zoom Phone user.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:user:admin`",
			"`phone:write:calling_plan:admin`",
			"`phone:delete:users_calling_plan:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "The ID of the Zoom user.",
			},
			"calling_plans": schema.SetNestedAttribute{
				Required:            true,
				MarkdownDescription: "Use this attribute to configure settings for the calling plan of the user.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.OneOf(lo.Values(callingPlanMapping)...),
							},
							MarkdownDescription: "The [type](https://marketplace.zoom.us/docs/api-reference/other-references/plans#zoom-phone-calling-plans) of calling plan. Allowed: " + strings.Join(lo.MapToSlice(callingPlanMapping, func(_ string, v int32) string { return fmt.Sprintf("`%d`", v) }), ", ") +
								strings.Join(
									append([]string{""}, lo.MapToSlice(callingPlanMapping, func(k string, v int32) string { return fmt.Sprintf("  - `%d`: %s", v, k) })...),
									markdownSeparatorForList,
								),
						},
						"billing_account_id": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The billing account ID. If the user is located in India, the field is required.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the calling plan.",
						},
					},
				},
			},
		},
	}
}

type resourceModel struct {
	UserID       types.String               `tfsdk:"user_id"`
	CallingPlans []resourceModelCallingPlan `tfsdk:"calling_plans"`
}

type resourceModelCallingPlan struct {
	Type             types.Int32  `tfsdk:"type"`
	BillingAccountID types.String `tfsdk:"billing_account_id"`
	Name             types.String `tfsdk:"name"`
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
		resp.Diagnostics.AddError(
			"Error reading phone calling plan of the user",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &output)
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
		UserID: plan.UserID,
		CallingPlans: lo.Map(dto.callingPlans, func(v readDtoCallingPlan, _ int) resourceModelCallingPlan {
			return resourceModelCallingPlan{
				Type:             v.callingPlanType,
				BillingAccountID: v.billingAccountID,
				Name:             v.callingPlanName,
			}
		}),
	}, nil
}

func (r *tfResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.create(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone calling plan of the user",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading phone calling plan of the user on creating",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, output)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *tfResource) create(ctx context.Context, plan resourceModel) error {
	_, err := r.crud.create(ctx, createDto{
		userID: plan.UserID,
		callingPlans: lo.Map(plan.CallingPlans, func(v resourceModelCallingPlan, _ int) createDtoCallingPlan {
			return createDtoCallingPlan{
				callingPlanType:  v.Type,
				billingAccountID: v.BillingAccountID,
			}
		}),
	})

	return err
}

func (r *tfResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.delete(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone calling plan of the user on updating",
			err.Error(),
		)
		return
	}

	if err := r.create(ctx, plan); err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone calling plan of the user on updating",
			err.Error(),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading phone calling plan of the user on reading",
			err.Error(),
		)
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

	if err := r.delete(ctx, state); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone calling plan of the user",
			fmt.Sprintf(
				"Could not delete phone calling plan of the user %s, unexpected error: %s",
				state.UserID.ValueString(),
				err,
			),
		)
		return
	}
}

func (r *tfResource) delete(ctx context.Context, state resourceModel) error {
	return r.crud.delete(ctx, deleteDto{
		userID: state.UserID,
		callingPlans: lo.Map(state.CallingPlans, func(v resourceModelCallingPlan, _ int) deleteDtoCallingPlan {
			return deleteDtoCallingPlan{
				callingPlanType:  v.Type,
				billingAccountID: v.BillingAccountID,
			}
		}),
	})
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("user_id"), req, resp)
}
