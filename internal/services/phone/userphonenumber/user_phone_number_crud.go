package userphonenumber

import (
	"context"
	"errors"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

func newCrud(client *zoomphone.Client) *crud {
	return &crud{
		client: client,
	}
}

type crud struct {
	client *zoomphone.Client
}

func (c *crud) read(ctx context.Context, userID types.String) (*readDto, error) {
	ret, err := c.client.PhoneUser(ctx, zoomphone.PhoneUserParams{
		UserId: userID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 404 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone user: %v", err)
	}
	phoneNumbers := lo.Map(ret.PhoneNumbers, func(p zoomphone.PhoneUserOKPhoneNumbersItem, _index int) *readDtoPhoneNumber {
		return &readDtoPhoneNumber{
			id:     util.FromOptString(p.ID),
			number: util.FromOptString(p.Number),
		}
	})
	return &readDto{
		phoneNumbers: phoneNumbers,
	}, nil
}

func (c *crud) assign(ctx context.Context, dto *assignDto) error {
	// Only a max of 5 numbers can be assigned to a call queue at a time.
	for _, phoneNumberIDs := range lo.Chunk(dto.phoneNumberIDs, 5) {
		_, err := c.client.AssignPhoneNumber(ctx, zoomphone.NewOptAssignPhoneNumberReq(
			zoomphone.AssignPhoneNumberReq{
				PhoneNumbers: lo.Map(phoneNumberIDs, func(phoneNumberID types.String, index int) zoomphone.AssignPhoneNumberReqPhoneNumbersItem {
					return zoomphone.AssignPhoneNumberReqPhoneNumbersItem{
						ID: util.ToOptString(phoneNumberID),
					}
				}),
			},
		), zoomphone.AssignPhoneNumberParams{UserId: dto.userID.ValueString()})
		if err != nil {
			return fmt.Errorf("error assigning phone user phone numbers by phone number id: %v", err)
		}
	}
	for _, phoneNumbers := range lo.Chunk(dto.phoneNumbers, 5) {
		_, err := c.client.AssignPhoneNumber(ctx, zoomphone.NewOptAssignPhoneNumberReq(
			zoomphone.AssignPhoneNumberReq{
				PhoneNumbers: lo.Map(phoneNumbers, func(phoneNumber types.String, index int) zoomphone.AssignPhoneNumberReqPhoneNumbersItem {
					return zoomphone.AssignPhoneNumberReqPhoneNumbersItem{
						Number: util.ToOptString(phoneNumber),
					}
				}),
			},
		), zoomphone.AssignPhoneNumberParams{UserId: dto.userID.ValueString()})
		if err != nil {
			return fmt.Errorf("error assigning phone user phone numbers by phone number: %v", err)
		}
	}
	return nil
}

func (c *crud) unassign(ctx context.Context, dto *unassignDto) error {
	for _, phoneNumberID := range dto.phoneNumberIDs {
		err := c.client.UnassignPhoneNumber(ctx, zoomphone.UnassignPhoneNumberParams{
			UserId:        dto.userID.ValueString(),
			PhoneNumberId: phoneNumberID.ValueString(),
		})
		if err != nil {
			var status *zoomphone.ErrorResponseStatusCode
			if errors.As(err, &status) {
				if status.StatusCode == 404 {
					continue
				}
			}
			return fmt.Errorf("error unassigning phone user phone numbers: %v", err)
		}
	}
	return nil
}
