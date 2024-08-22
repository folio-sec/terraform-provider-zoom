package callqueuephonenumber

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

func (c *crud) read(ctx context.Context, callQueueID types.String) (*readDto, error) {
	ret, err := c.client.GetACallQueue(ctx, zoomphone.GetACallQueueParams{
		CallQueueId: callQueueID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone call queue phone numbers: %v", err)
	}
	phoneNumbers := lo.Map(ret.PhoneNumbers, func(p zoomphone.GetACallQueueOKPhoneNumbersItem, _index int) *readDtoPhoneNumber {
		return &readDtoPhoneNumber{
			id:     util.FromOptString(p.ID),
			number: util.FromOptString(p.Number),
			source: util.FromOptString(p.Source),
		}
	})
	return &readDto{
		phoneNumbers: phoneNumbers,
	}, nil
}

func (c *crud) assign(ctx context.Context, dto *assignDto) error {
	// Only a max of 5 numbers can be assigned to a call queue at a time.
	for _, phoneNumberIDs := range lo.Chunk(dto.phoneNumberIDs, 5) {
		err := c.client.AssignPhoneToCallQueue(ctx, zoomphone.NewOptAssignPhoneToCallQueueReq(
			zoomphone.AssignPhoneToCallQueueReq{
				PhoneNumbers: lo.Map(phoneNumberIDs, func(phoneNumberID types.String, index int) zoomphone.AssignPhoneToCallQueueReqPhoneNumbersItem {
					return zoomphone.AssignPhoneToCallQueueReqPhoneNumbersItem{
						ID: util.ToOptString(phoneNumberID),
					}
				}),
			},
		), zoomphone.AssignPhoneToCallQueueParams{CallQueueId: dto.callQueueID.ValueString()})
		if err != nil {
			return fmt.Errorf("error assigning phone call queue phone numbers by phone number id: %v", err)
		}
	}
	for _, phoneNumbers := range lo.Chunk(dto.phoneNumbers, 5) {
		err := c.client.AssignPhoneToCallQueue(ctx, zoomphone.NewOptAssignPhoneToCallQueueReq(
			zoomphone.AssignPhoneToCallQueueReq{
				PhoneNumbers: lo.Map(phoneNumbers, func(phoneNumber types.String, index int) zoomphone.AssignPhoneToCallQueueReqPhoneNumbersItem {
					return zoomphone.AssignPhoneToCallQueueReqPhoneNumbersItem{
						Number: util.ToOptString(phoneNumber),
					}
				}),
			},
		), zoomphone.AssignPhoneToCallQueueParams{CallQueueId: dto.callQueueID.ValueString()})
		if err != nil {
			return fmt.Errorf("error assigning phone call queue phone numbers by phone number: %v", err)
		}
	}
	return nil
}

func (c *crud) unassign(ctx context.Context, dto *unassignDto) error {
	for _, phoneNumberID := range dto.phoneNumberIDs {
		err := c.client.UnAssignPhoneNumCallQueue(ctx, zoomphone.UnAssignPhoneNumCallQueueParams{
			CallQueueId:   dto.callQueueID.ValueString(),
			PhoneNumberId: phoneNumberID.ValueString(),
		})
		if err != nil {
			var status *zoomphone.ErrorResponseStatusCode
			if errors.As(err, &status) {
				if status.StatusCode == 404 {
					continue
				}
			}
			return fmt.Errorf("error unassigning phone call queue phone numbers: %v", err)
		}
	}
	return nil
}

func (c *crud) unassignAll(ctx context.Context, callQueueID types.String) error {
	err := c.client.UnassignAPhoneNumCallQueue(ctx, zoomphone.UnassignAPhoneNumCallQueueParams{
		CallQueueId: callQueueID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 404 {
				return nil
			}
		}
		return fmt.Errorf("error unassigning all phone call queue phone numbers: %v", err)
	}
	return nil
}
