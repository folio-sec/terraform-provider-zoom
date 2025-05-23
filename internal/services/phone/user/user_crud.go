package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomuser"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

func newCrud(phoneClient *zoomphone.Client, userClient *zoomuser.Client) *crud {
	return &crud{
		phoneClient: phoneClient,
		userClient:  userClient,
	}
}

type crud struct {
	phoneClient *zoomphone.Client
	userClient  *zoomuser.Client
}

func (c *crud) list(ctx context.Context, dto listQueryDto) (*listDto, error) {
	var users []listDtoUser
	nextPageToken := zoomphone.OptString{}

	for {
		ret, err := c.phoneClient.ListPhoneUsers(ctx, zoomphone.ListPhoneUsersParams{
			PageSize:      zoomphone.NewOptInt(100), // Max 100
			NextPageToken: nextPageToken,
			SiteID:        util.ToPhoneOptString(dto.siteID),
			CallingType:   util.ToPhoneOptInt(dto.callingType),
			Status:        util.ToPhoneOptString(dto.status),
			Department:    util.ToPhoneOptString(dto.department),
			CostCenter:    util.ToPhoneOptString(dto.costCenter),
			Keyword:       util.ToPhoneOptString(dto.keyword),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to list users: %v", err)
		}

		users = append(users, lo.Map(ret.Users, func(item zoomphone.ListPhoneUsersOKUsersItem, _ int) listDtoUser {
			return listDtoUser{
				callingPlans: lo.Map(item.CallingPlans, func(item zoomphone.ListPhoneUsersOKUsersItemCallingPlansItem, index int) *listDtoUserCallingPlan {
					return &listDtoUserCallingPlan{
						name:               util.FromOptString(item.Name),
						typ:                util.FromOptInt(item.Type),
						billingAccountID:   util.FromOptString(item.BillingAccountID),
						billingAccountName: util.FromOptString(item.BillingAccountName),
					}
				}),
				email:           util.FromOptString(item.Email),
				extensionID:     util.FromOptString(item.ExtensionID),
				extensionNumber: util.FromOptInt64(item.ExtensionNumber),
				userID:          util.FromOptString(item.ID),
				name:            util.FromOptString(item.Name),
				phoneUserID:     util.FromOptString(item.PhoneUserID),
				site: lo.TernaryF(item.Site.IsSet(), func() *listDtoUserSite {
					return &listDtoUserSite{
						id:   util.FromOptString(item.Site.Value.ID),
						name: util.FromOptString(item.Site.Value.Name),
					}
				}, func() *listDtoUserSite {
					return nil
				}),
				status: util.FromOptString(item.Status),
				phoneNumbers: lo.Map(item.PhoneNumbers, func(item zoomphone.ListPhoneUsersOKUsersItemPhoneNumbersItem, index int) *listDtoUserPhoneNumber {
					return &listDtoUserPhoneNumber{
						id:     util.FromOptString(item.ID),
						number: util.FromOptString(item.Number),
					}
				}),
				department: util.FromOptString(item.Department),
				costCenter: util.FromOptString(item.CostCenter),
			}
		})...)

		if !ret.NextPageToken.IsSet() || ret.NextPageToken.Value == "" {
			break
		}
		nextPageToken = ret.NextPageToken
	}

	return &listDto{
		// Ensure uniqueness by using user ID, as duplicate data may occasionally be retrieved.
		users: lo.UniqBy(users, func(item listDtoUser) string {
			return item.userID.ValueString()
		}),
	}, nil
}

func (c *crud) read(ctx context.Context, zoomUserID types.String) (*readDto, error) {
	detail, err := c.phoneClient.PhoneUser(ctx, zoomphone.PhoneUserParams{
		UserId: zoomUserID.ValueString(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to read phone user: %v", err)
	}

	return &readDto{
		callingPlans: lo.Map(detail.CallingPlans, func(callingPlan zoomphone.PhoneUserOKCallingPlansItem, _ int) readDtoCallingPlan {
			return readDtoCallingPlan{
				callingPlanType:    util.FromOptInt(callingPlan.Type),
				billingAccountID:   util.FromOptString(callingPlan.BillingAccountID),
				billingAccountName: util.FromOptString(callingPlan.BillingAccountName),
			}
		}),
		costCenter:         util.FromOptString(detail.CostCenter),
		department:         util.FromOptString(detail.Department),
		email:              util.FromOptString(detail.Email),
		emergencyAddressID: util.FromOptString(detail.EmergencyAddress.Value.ID),
		extensionID:        util.FromOptString(detail.ExtensionID),
		extensionNumber:    util.FromOptInt64(detail.ExtensionNumber),
		zoomUserID:         util.FromOptString(detail.ID),
		phoneNumbers: lo.Map(detail.PhoneNumbers, func(phoneNumber zoomphone.PhoneUserOKPhoneNumbersItem, _ int) readDtoPhoneNumber {
			return readDtoPhoneNumber{
				phoneNumberID: util.FromOptString(phoneNumber.ID),
				phoneNumber:   util.FromOptString(phoneNumber.Number),
			}
		}),
		phoneUserID: util.FromOptString(detail.PhoneUserID),
		siteID:      util.FromOptString(detail.SiteID),
		status:      util.FromOptString(detail.Status),
	}, nil
}

func (c *crud) create(ctx context.Context, dto createDto) (*createdDto, error) {
	// There is no API to create a zoom phone user.
	// Using the behavior that a phone user is created by changing the feature.zoom_phone attribute of the zoom user.
	err := c.userClient.UserUpdate(ctx, zoomuser.NewOptUserUpdateReq(zoomuser.UserUpdateReq{
		Feature: zoomuser.NewOptUserUpdateReqFeature(zoomuser.UserUpdateReqFeature{
			ZoomPhone: zoomuser.NewOptBool(true),
		}),
	}), zoomuser.UserUpdateParams{
		UserId: dto.zoomUserID.ValueString(),
	})

	if err != nil {
		return nil, fmt.Errorf("error creating phone user: %v", err)
	}

	return &createdDto{}, nil
}

func (c *crud) update(ctx context.Context, dto updateDto) error {
	err := c.phoneClient.UpdateUserProfile(ctx, zoomphone.NewOptUpdateUserProfileReq(
		zoomphone.UpdateUserProfileReq{
			EmergencyAddressID: util.ToPhoneOptString(dto.emergencyAddressID),
			ExtensionNumber: lo.TernaryF(dto.extensionNumber.IsNull() || dto.extensionNumber.IsUnknown(), func() zoomphone.OptString {
				return zoomphone.OptString{}
			}, func() zoomphone.OptString {
				return zoomphone.NewOptString(strconv.FormatInt(dto.extensionNumber.ValueInt64(), 10))
			}),
			SiteID:     util.ToPhoneOptString(dto.siteID),
			TemplateID: util.ToPhoneOptString(dto.templateID),
		},
	), zoomphone.UpdateUserProfileParams{
		UserId: dto.zoomUserID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone user: %v", err)
	}

	return nil
}

func (c *crud) delete(ctx context.Context, zoomUserID types.String) error {
	// There is no API to delete a zoom phone user.
	// Using the behavior that a phone user is deleted by changing the feature.zoom_phone attribute of the zoom user.
	err := c.userClient.UserUpdate(ctx, zoomuser.NewOptUserUpdateReq(zoomuser.UserUpdateReq{
		Feature: zoomuser.NewOptUserUpdateReqFeature(zoomuser.UserUpdateReqFeature{
			ZoomPhone: zoomuser.NewOptBool(false),
		}),
	}), zoomuser.UserUpdateParams{
		UserId: zoomUserID.ValueString(),
	})

	if err != nil {
		var status *zoomuser.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.Response.Code.Value == http.StatusNotFound {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone user: %v", err)
	}

	return nil
}
