package sharedlinegroupgroupphonenumbers

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

func (c *crud) read(ctx context.Context, sharedLineGroupID types.String) (*readDto, error) {
	ret, err := c.client.GetASharedLineGroup(ctx, zoomphone.GetASharedLineGroupParams{
		SharedLineGroupId: sharedLineGroupID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone shared line group phone numbers: %v", err)
	}
	phoneNumbers := lo.Map(ret.PhoneNumbers, func(p zoomphone.GetASharedLineGroupOKPhoneNumbersItem, _index int) *readDtoPhoneNumber {
		return &readDtoPhoneNumber{
			id:     util.FromOptString(p.ID),
			number: util.FromOptString(p.Number),
		}
	})
	return &readDto{
		primaryNumber: util.FromOptString(ret.PrimaryNumber),
		phoneNumbers:  phoneNumbers,
	}, nil
}

func (c *crud) assign(ctx context.Context, dto *assignDto) error {
	// Only a max of 5 numbers can be assigned to a shared line group at a time.
	for _, phoneNumberIDs := range lo.Chunk(dto.phoneNumberIDs, 5) {
		err := c.client.AssignPhoneNumbersSLG(ctx, zoomphone.NewOptAssignPhoneNumbersSLGReq(
			zoomphone.AssignPhoneNumbersSLGReq{
				PhoneNumbers: lo.Map(phoneNumberIDs, func(phoneNumberID types.String, index int) zoomphone.AssignPhoneNumbersSLGReqPhoneNumbersItem {
					return zoomphone.AssignPhoneNumbersSLGReqPhoneNumbersItem{
						ID: util.ToOptString(phoneNumberID),
					}
				}),
			},
		), zoomphone.AssignPhoneNumbersSLGParams{SharedLineGroupId: dto.sharedLineGroupID.ValueString()})
		if err != nil {
			return fmt.Errorf("error assigning phone shared line group phone numbers by phone number id: %v", err)
		}
	}
	for _, phoneNumbers := range lo.Chunk(dto.phoneNumbers, 5) {
		err := c.client.AssignPhoneNumbersSLG(ctx, zoomphone.NewOptAssignPhoneNumbersSLGReq(
			zoomphone.AssignPhoneNumbersSLGReq{
				PhoneNumbers: lo.Map(phoneNumbers, func(phoneNumber types.String, index int) zoomphone.AssignPhoneNumbersSLGReqPhoneNumbersItem {
					return zoomphone.AssignPhoneNumbersSLGReqPhoneNumbersItem{
						Number: util.ToOptString(phoneNumber),
					}
				}),
			},
		), zoomphone.AssignPhoneNumbersSLGParams{SharedLineGroupId: dto.sharedLineGroupID.ValueString()})
		if err != nil {
			return fmt.Errorf("error assigning phone shared line group phone numbers by phone number: %v", err)
		}
	}
	return nil
}

func (c *crud) unassign(ctx context.Context, dto *unassignDto) error {
	for _, phoneNumberID := range dto.phoneNumberIDs {
		err := c.client.DeleteAPhoneNumberSLG(ctx, zoomphone.DeleteAPhoneNumberSLGParams{
			SharedLineGroupId: dto.sharedLineGroupID.ValueString(),
			PhoneNumberId:     phoneNumberID.ValueString(),
		})
		if err != nil {
			var status *zoomphone.ErrorResponseStatusCode
			if errors.As(err, &status) {
				if status.StatusCode == 404 {
					continue
				}
			}
			return fmt.Errorf("error unassigning phone shared line group phone numbers: %v", err)
		}
	}
	return nil
}

func (c *crud) unassignAll(ctx context.Context, sharedLineGroupID types.String) error {
	err := c.client.DeletePhoneNumbersSLG(ctx, zoomphone.DeletePhoneNumbersSLGParams{
		SharedLineGroupId: sharedLineGroupID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 404 {
				return nil
			}
		}
		return fmt.Errorf("error unassigning all phone shared line group phone numbers: %v", err)
	}
	return nil
}

func (c *crud) updatePrimaryNumber(ctx context.Context, sharedLineGroupID, primaryNumber types.String) error {
	// UpdateASharedLineGroup is PATCH, so just update PrimaryNumber only
	err := c.client.UpdateASharedLineGroup(ctx, zoomphone.OptUpdateASharedLineGroupReq{
		Value: zoomphone.UpdateASharedLineGroupReq{
			PrimaryNumber: util.ToOptString(primaryNumber),
		},
		Set: true,
	}, zoomphone.UpdateASharedLineGroupParams{
		SharedLineGroupId: sharedLineGroupID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone shared line group primary number: %v", err)
	}
	return nil
}
