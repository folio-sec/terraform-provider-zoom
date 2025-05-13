package site

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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

func NewPhoneSiteResource() resource.Resource {
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
	resp.TypeName = req.ProviderTypeName + "_phone_site"
}

func (r *tfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a site within Zoom Phone.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:site:admin`",
			"`phone:write:site:admin`",
			"`phone:update:site:admin`",
			"`phone:delete:site:admin`",
			"`phone:read:list_sites:admin`",
			"`phone:read:list_emergency_addresses:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The site ID is the unique identifier of the site.",
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
				MarkdownDescription: "The name of the site. Constraints: Max 255 chars.",
			},
			"main_auto_receptionist": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "The [main auto receptionist](https://support.zoom.us/hc/en-us/articles/360021121312#h_bc7ff1d5-0e6c-40cd-b889-62010cb98c57) for each site.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "The auto receptionist ID.",
					},
					"name": schema.StringAttribute{
						Required:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						MarkdownDescription: "Display name of the [auto-receptionist](https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0061421) as main auto-receptionist for the site.",
					},
				},
			},
			"source_auto_receptionist_id": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "The ID of the [auto-receptionist](https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0061421) that can be copied when only creating as main auto-receptionist.",
			},
			"default_emergency_address": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "The default [emergency address](https://support.zoom.us/hc/en-us/articles/360021062871-Setting-an-Emergency-Address) at creation time. If the address provided is not an exact match, it uses the system generated corrected address.",
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"address_line1": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The address Line 1 of the emergency address that contains the house number and street name.",
					},
					"address_line2": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The address Line 2 of the emergency address that contains the building number, floor number, unit, and others.",
					},
					"city": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The city of the emergency address.",
					},
					"country": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(2),
							stringvalidator.LengthAtMost(2),
						},
						MarkdownDescription: "The two-lettered country code (Alpha-2 code in ISO-3166 format) of the site's emergency address.",
					},
					"state_code": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The state code of the emergency address.",
					},
					"zip": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The ZIP code of the emergency address.",
					},
				},
			},
			"site_code": schema.Int32Attribute{
				Optional: true,
				Validators: []validator.Int32{
					int32validator.Between(1, 999999),
				},
				MarkdownDescription: "The [site code](https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0069806).",
			},
			"short_extension": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The short extension of the phone site.",
				Attributes: map[string]schema.Attribute{
					"length": schema.Int32Attribute{
						Optional: true,
						Computed: true,
						Validators: []validator.Int32{
							int32validator.Between(1, 6),
						},
						MarkdownDescription: "This setting specifies the length of short extension numbers for the site. The value must be between 1 and 6., Default is `3`.",
					},
					"ranges": schema.SetNestedAttribute{
						Optional:            true,
						MarkdownDescription: "The range list. After adding a short extension range, the newly assigned extension numbers start from the `range_from` value. ",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"range_from": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(regexp.MustCompile("^[1-9]?[0-9]{1,5}$"), "value must be 1 to 6 digits"),
									},
									MarkdownDescription: "The short extension's starting range number, which can be a non-negative value. This value must be less than the `range_to` value.",
								},
								"range_to": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.RegexMatches(regexp.MustCompile("^[1-9]?[0-9]{1,5}$"), "value must be 1 to 6 digits"),
									},
									MarkdownDescription: "The short extension's ending range number, which can be a non-negative value. This value cannot be less than or equal to the `range_from` value.",
								},
							},
						},
					},
				},
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						map[string]attr.Type{
							"length": types.Int32Type,
							"ranges": types.SetType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
								"range_from": types.StringType,
								"range_to":   types.StringType,
							}}},
						},
						map[string]attr.Value{
							"length": types.Int32Value(3),
							"ranges": types.SetNull(types.ObjectType{AttrTypes: map[string]attr.Type{
								"range_from": types.StringType,
								"range_to":   types.StringType,
							}}),
						},
					),
				),
			},
			"sip_zone_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The SIP zone ID. If the account enabled the Display Custom SIP Zone Options on Web Portal feature, then selecting a SIP zone nearest to your site might help reduce latency and improve call quality.",
			},
			"caller_id_name": schema.StringAttribute{
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtMost(15)},
				MarkdownDescription: "When an outbound call uses a number as the caller ID, the caller ID name and the number display to the called party. The caller ID name can be up to 15 characters. The user can reset the caller ID name by setting it to empty string.",
			},
			"level": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The level of the site.",
			},
			"india_state_code": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
				MarkdownDescription: "The India site's state code. This field only applies to India based accounts.",
			},
			"india_city": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
				MarkdownDescription: "The India site's city. This field only applies to India based accounts.",
			},
			"india_sdca_npa": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
				MarkdownDescription: "The India site's Short Distance Calling Area (sdca) Numbering Plan Area (npa). This field is linked to the 'state_code' field. This field only applies to India based accounts.",
			},
			"india_entity_name": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplaceIfConfigured()},
				MarkdownDescription: "When select the Indian sip zone, then need to set the entity name. This field only applies to India based accounts.",
			},
		},
	}
}

type resourceModel struct {
	ID                       types.String                          `tfsdk:"id"`
	Name                     types.String                          `tfsdk:"name"`
	MainAutoReceptionist     *resourceModelMainAutoReceptionist    `tfsdk:"main_auto_receptionist"`
	SourceAutoReceptionistID types.String                          `tfsdk:"source_auto_receptionist_id"`
	DefaultEmergencyAddress  *resourceModelDefaultEmergencyAddress `tfsdk:"default_emergency_address"`
	SiteCode                 types.Int32                           `tfsdk:"site_code"`
	ShortExtension           *resourceModelShortExtension          `tfsdk:"short_extension"`
	SipZoneID                types.String                          `tfsdk:"sip_zone_id"`
	CallerIDName             types.String                          `tfsdk:"caller_id_name"`
	Level                    types.String                          `tfsdk:"level"`
	IndiaStateCode           types.String                          `tfsdk:"india_state_code"`
	IndiaCity                types.String                          `tfsdk:"india_city"`
	IndiaSdcaNpa             types.String                          `tfsdk:"india_sdca_npa"`
	IndiaEntityName          types.String                          `tfsdk:"india_entity_name"`
}

type resourceModelMainAutoReceptionist struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type resourceModelDefaultEmergencyAddress struct {
	AddressLine1 types.String `tfsdk:"address_line1"`
	AddressLine2 types.String `tfsdk:"address_line2"`
	City         types.String `tfsdk:"city"`
	Country      types.String `tfsdk:"country"`
	StateCode    types.String `tfsdk:"state_code"`
	Zip          types.String `tfsdk:"zip"`
}

type resourceModelShortExtension struct {
	Length types.Int32                         `tfsdk:"length"`
	Ranges *[]resourceModelShortExtensionRange `tfsdk:"ranges"`
}

type resourceModelShortExtensionRange struct {
	RangeFrom types.String `tfsdk:"range_from"`
	RangeTo   types.String `tfsdk:"range_to"`
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
		resp.Diagnostics.AddError("Error reading phone site", err.Error())
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
		return nil, fmt.Errorf("error read phone site: %v", err)
	}
	if dto == nil {
		return nil, nil // already deleted
	}

	return &resourceModel{
		ID:   dto.id,
		Name: dto.name,
		MainAutoReceptionist: &resourceModelMainAutoReceptionist{
			ID:   dto.mainAutoReceptionist.id,
			Name: dto.mainAutoReceptionist.name,
		},
		// SourceAutoReceptionistID is used only at the time of creation and cannot be read, so use the value of plan.
		SourceAutoReceptionistID: plan.SourceAutoReceptionistID,
		// DefaultEmergencyAddress is used only at the time of creation and cannot be read, so use the value of plan.
		DefaultEmergencyAddress: lo.TernaryF(plan.DefaultEmergencyAddress != nil, func() *resourceModelDefaultEmergencyAddress {
			return &resourceModelDefaultEmergencyAddress{
				AddressLine1: plan.DefaultEmergencyAddress.AddressLine1,
				AddressLine2: plan.DefaultEmergencyAddress.AddressLine2,
				City:         plan.DefaultEmergencyAddress.City,
				Country:      plan.DefaultEmergencyAddress.Country,
				StateCode:    plan.DefaultEmergencyAddress.StateCode,
				Zip:          plan.DefaultEmergencyAddress.Zip,
			}
		}, lo.Nil),
		SiteCode: dto.siteCode,
		ShortExtension: lo.TernaryF(!dto.shortExtensionLength.IsNull(), func() *resourceModelShortExtension {
			return &resourceModelShortExtension{
				Length: dto.shortExtensionLength,
				// Ranges are used only at the time of creation and update, and cannot be read, so use the value of plan.
				Ranges: lo.TernaryF(plan.ShortExtension != nil, func() *[]resourceModelShortExtensionRange {
					return plan.ShortExtension.Ranges
				}, lo.Nil),
			}
		}, func() *resourceModelShortExtension {
			return &resourceModelShortExtension{
				// The API response may not include the default value 3, so we explicitly set it to 3 when the value is null.
				Length: types.Int32Value(3),
				// Ranges are used only at the time of creation and update, and cannot be read, so use the value of plan.
				Ranges: lo.TernaryF(plan.ShortExtension != nil, func() *[]resourceModelShortExtensionRange {
					return plan.ShortExtension.Ranges
				}, lo.Nil),
			}
		}),
		SipZoneID:       dto.sipZone.id,
		CallerIDName:    dto.callerIDName,
		Level:           dto.level,
		IndiaStateCode:  lo.Ternary(dto.indiaStateCode.ValueString() != "", dto.indiaStateCode, types.StringNull()),
		IndiaCity:       lo.Ternary(dto.indiaCity.ValueString() != "", dto.indiaCity, types.StringNull()),
		IndiaSdcaNpa:    lo.Ternary(dto.indiaSdcaNpa.ValueString() != "", dto.indiaSdcaNpa, types.StringNull()),
		IndiaEntityName: lo.Ternary(dto.indiaEntityName.ValueString() != "", dto.indiaEntityName, types.StringNull()),
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
		autoReceptionistName:     plan.MainAutoReceptionist.Name,
		sourceAutoReceptionistID: plan.SourceAutoReceptionistID,
		defaultEmergencyAddress: createDtoDefaultEmergencyAddress{
			addressLine1: plan.DefaultEmergencyAddress.AddressLine1,
			addressLine2: plan.DefaultEmergencyAddress.AddressLine2,
			city:         plan.DefaultEmergencyAddress.City,
			stateCode:    plan.DefaultEmergencyAddress.StateCode,
			countryCode:  plan.DefaultEmergencyAddress.Country,
			zip:          plan.DefaultEmergencyAddress.Zip,
		},
		name: plan.Name,
		shortExtensionLength: lo.TernaryF(plan.ShortExtension != nil, func() types.Int32 {
			return plan.ShortExtension.Length
		}, func() types.Int32 {
			return types.Int32Null()
		}),
		siteCode:        plan.SiteCode,
		sipZoneID:       plan.SipZoneID,
		indiaStateCode:  plan.IndiaStateCode,
		indiaCity:       plan.IndiaCity,
		indiaSdcaNpa:    plan.IndiaSdcaNpa,
		indiaEntityName: plan.IndiaEntityName,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating phone site",
			err.Error(),
		)
		return
	}

	// some fields cannot set on create api, so update them
	if err := r.crud.update(ctx, &updateDto{
		id:       ret.id,
		name:     ret.name,
		siteCode: plan.SiteCode,
		shortExtension: lo.TernaryF(plan.ShortExtension != nil, func() *updateDtoShortExtension {
			return &updateDtoShortExtension{
				length: plan.ShortExtension.Length,
				ranges: lo.TernaryF(plan.ShortExtension.Ranges != nil, func() []updateDtoShortExtensionRange {
					return lo.Map(*plan.ShortExtension.Ranges, func(item resourceModelShortExtensionRange, _ int) updateDtoShortExtensionRange {
						return updateDtoShortExtensionRange{
							rangeFrom: item.RangeFrom,
							rangeTo:   item.RangeTo,
						}
					})
				}, func() []updateDtoShortExtensionRange {
					return []updateDtoShortExtensionRange{}
				}),
			}
		}, lo.Nil),
		sipZoneID:    plan.SipZoneID,
		callerIDName: plan.CallerIDName,
	}); err != nil {
		_ = r.delete(ctx, ret.id)
		resp.Diagnostics.AddError(
			"Error creating phone site on updating",
			err.Error(),
		)
		return
	}

	// Since the ID is unknown at the time of creation, set the retrieved value.
	plan.ID = ret.id
	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error creating phone site on reading", err.Error())
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
			"Error updating phone site on get plan",
			"Error updating phone site",
		)
		return
	}

	if err := r.crud.update(ctx, &updateDto{
		id:       plan.ID,
		name:     plan.Name,
		siteCode: plan.SiteCode,
		shortExtension: lo.TernaryF(plan.ShortExtension != nil, func() *updateDtoShortExtension {
			return &updateDtoShortExtension{
				length: plan.ShortExtension.Length,
				ranges: lo.TernaryF(plan.ShortExtension.Ranges != nil, func() []updateDtoShortExtensionRange {
					return lo.Map(*plan.ShortExtension.Ranges, func(item resourceModelShortExtensionRange, _ int) updateDtoShortExtensionRange {
						return updateDtoShortExtensionRange{
							rangeFrom: item.RangeFrom,
							rangeTo:   item.RangeTo,
						}
					})
				}, func() []updateDtoShortExtensionRange {
					return []updateDtoShortExtensionRange{}
				}),
			}
		}, lo.Nil),
		sipZoneID:    plan.SipZoneID,
		callerIDName: plan.CallerIDName,
	}); err != nil {
		resp.Diagnostics.AddError(
			"Error updating phone site",
			fmt.Sprintf(
				"Could not update phone site %s, unexpected error: %s",
				plan.ID.ValueString(),
				err,
			),
		)
		return
	}

	output, err := r.read(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error updating phone site", err.Error())
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

	if err := r.delete(ctx, state.ID); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting phone site",
			fmt.Sprintf(
				"Could not delete phone site %s, unexpected error: %s",
				state.ID.ValueString(),
				err,
			),
		)
		return
	}

	tflog.Info(ctx, "deleted phone site", map[string]interface{}{
		"site_id": state.ID.ValueString(),
	})
}

func (r *tfResource) delete(ctx context.Context, siteId types.String) error {
	mainSite, err := r.crud.readMain(ctx)
	if err != nil {
		return fmt.Errorf("error read phone main site: %v", err)
	}
	return r.crud.delete(ctx, siteId, mainSite.id)
}

func (r *tfResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
