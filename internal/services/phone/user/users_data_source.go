package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ datasource.DataSource              = &tfDataSource{}
	_ datasource.DataSourceWithConfigure = &tfDataSource{}
)

func NewPhoneUsersDataSource() datasource.DataSource {
	return &tfDataSource{}
}

type tfDataSource struct {
	crud *crud
}

func (d *tfDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.crud = newCrud(data.PhoneClient, data.UserClient)
}

func (d *tfDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_users"
}

func (d *tfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `A list of all of an account's users who are assigned a Zoom Phone license.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the ` + strings.Join([]string{
			"`phone:read:list_users:admin`",
		}, ", ") + ".",
		Attributes: map[string]schema.Attribute{
			"query": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "The query parameters for listing users.",
				Attributes: map[string]schema.Attribute{
					"site_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The unique identifier of the site. Get it from the [List Phone Sites](https://marketplace.zoom.us/docs/api-reference/phone/methods#operation/listPhoneSites) API.",
					},
					"calling_type": schema.Int32Attribute{
						Optional:            true,
						MarkdownDescription: "The [type](https://marketplace.zoom.us/docs/api-reference/other-references/plans#zoom-phone-calling-plans) of calling plan.",
					},
					"status": schema.StringAttribute{
						Optional: true,
						MarkdownDescription: `The status of the Zoom Phone user.
  - pending: The users have been assigned the Zoom Workplace license, but not been assigned Zoom Phone feature.
  - Allowed: activate ┃ deactivate ┃ pending
`,
					},
					"department": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The department where the user belongs.",
					},
					"cost_center": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The cost center where the user belongs.",
					},
					"keyword": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The partial string of user's name, extension number, or phone number.",
					},
				},
			},
			"users": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "A list of users. Each user object provides the attributes documented below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the Zoom user.",
						},
						"phone_user_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the Phone user.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the user.",
						},
						"email": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The email address of the user.",
						},
						"extension_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The extension ID.",
						},
						"extension_number": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The extension number assigned to the user's Zoom phone number.",
						},
						"status": schema.StringAttribute{
							Computed: true,
							MarkdownDescription: `The status of the user's Zoom Phone license. The value can be either of the following:
  - activate: Active Zoom phone user.
  - deactivate: User with Zoom phone license disabled. This type of user can't make or receive calls.
`,
						},
						"department": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The department where the user belongs.",
						},
						"cost_center": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The cost center where the user belongs.",
						},
						"site_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The unique identifier of the [site](https://support.zoom.us/hc/en-us/articles/360020809672z) where the user should be moved or assigned.",
						},
						"phone_numbers": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The phone number ID.",
									},
									"number": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The phone number.",
									},
								},
							},
						},
						"calling_plans": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The name of the user's calling plan.",
									},
									"type": schema.Int32Attribute{
										Computed:            true,
										MarkdownDescription: "The type of calling plan where the user is enrolled.",
									},
									"billing_account_id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The billing account ID. It displays when the user is located in India.",
									},
									"billing_account_name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The billing account name. It displays when the user is located in India.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type dataSourceModel struct {
	Query *dataSourceModelQuery  `tfsdk:"query"`
	Users []*dataSourceModelUser `tfsdk:"users"`
}

type dataSourceModelQuery struct {
	SiteID      types.String `tfsdk:"site_id"`
	CallingType types.Int32  `tfsdk:"calling_type"`
	Status      types.String `tfsdk:"status"`
	Department  types.String `tfsdk:"department"`
	CostCenter  types.String `tfsdk:"cost_center"`
	Keyword     types.String `tfsdk:"keyword"`
}

type dataSourceModelUser struct {
	UserID          types.String                      `tfsdk:"user_id"`
	PhoneUserID     types.String                      `tfsdk:"phone_user_id"`
	Name            types.String                      `tfsdk:"name"`
	Email           types.String                      `tfsdk:"email"`
	ExtensionID     types.String                      `tfsdk:"extension_id"`
	ExtensionNumber types.Int64                       `tfsdk:"extension_number"`
	Status          types.String                      `tfsdk:"status"`
	Department      types.String                      `tfsdk:"department"`
	CostCenter      types.String                      `tfsdk:"cost_center"`
	SiteID          types.String                      `tfsdk:"site_id"`
	PhoneNumbers    []*dataSourceModelUserPhoneNumber `tfsdk:"phone_numbers"`
	CallingPlans    []*dataSourceModelUserCallingPlan `tfsdk:"calling_plans"`
}

type dataSourceModelUserPhoneNumber struct {
	ID     types.String `tfsdk:"id"`
	Number types.String `tfsdk:"number"`
}

type dataSourceModelUserCallingPlan struct {
	Name               types.String `tfsdk:"name"`
	Type               types.Int32  `tfsdk:"type"`
	BillingAccountID   types.String `tfsdk:"billing_account_id"`
	BillingAccountName types.String `tfsdk:"billing_account_name"`
}

func (d *tfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dto, err := d.crud.list(ctx, lo.TernaryF(data.Query == nil, func() listQueryDto {
		return listQueryDto{}
	}, func() listQueryDto {
		return listQueryDto{
			siteID:      data.Query.SiteID,
			callingType: data.Query.CallingType,
			status:      data.Query.Status,
			department:  data.Query.Department,
			costCenter:  data.Query.CostCenter,
			keyword:     data.Query.Keyword,
		}
	}))
	if err != nil {
		resp.Diagnostics.AddError("Error reading users", err.Error())
		return
	}

	data.Users = lo.Map(dto.users, func(user listDtoUser, _ int) *dataSourceModelUser {
		siteID := types.StringNull()
		if user.site != nil {
			siteID = user.site.id
		}
		return &dataSourceModelUser{
			UserID:          user.userID,
			PhoneUserID:     user.phoneUserID,
			Name:            user.name,
			Email:           user.email,
			ExtensionID:     user.extensionID,
			ExtensionNumber: user.extensionNumber,
			Status:          user.status,
			Department:      user.department,
			CostCenter:      user.costCenter,
			SiteID:          siteID,
			PhoneNumbers: lo.Map(user.phoneNumbers, func(item *listDtoUserPhoneNumber, index int) *dataSourceModelUserPhoneNumber {
				return &dataSourceModelUserPhoneNumber{
					ID:     item.id,
					Number: item.number,
				}
			}),
			CallingPlans: lo.Map(user.callingPlans, func(item *listDtoUserCallingPlan, index int) *dataSourceModelUserCallingPlan {
				return &dataSourceModelUserCallingPlan{
					Name:               item.name,
					Type:               item.typ,
					BillingAccountID:   item.billingAccountID,
					BillingAccountName: item.billingAccountName,
				}
			}),
		}
	})

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
