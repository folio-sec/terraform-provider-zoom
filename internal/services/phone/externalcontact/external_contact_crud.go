package externalcontact

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newCrud(client *zoomphone.Client) *crud {
	return &crud{
		client: client,
	}
}

type crud struct {
	client *zoomphone.Client
}

func (c *crud) read(ctx context.Context, externalContactID types.String) (*readDto, error) {
	detail, err := c.client.GetAExternalContact(ctx, zoomphone.GetAExternalContactParams{
		ExternalContactId: externalContactID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone external contact: %v", err)
	}

	return &readDto{
		description:       util.FromOptString(detail.Description),
		email:             util.FromOptString(detail.Email),
		extensionNumber:   util.FromOptString(detail.ExtensionNumber),
		externalContactID: util.FromOptString(detail.ExternalContactID),
		id:                util.FromOptString(detail.ID),
		name:              util.FromOptString(detail.Name),
		phoneNumbers: lo.Map(detail.PhoneNumbers, func(item string, index int) types.String {
			return types.StringValue(item)
		}),
		routingPath:      util.FromOptString(detail.RoutingPath),
		autoCallRecorded: util.FromOptBool(detail.AutoCallRecorded),
	}, nil
}

func (c *crud) create(ctx context.Context, dto *createDto) (*createdDto, error) {
	res, err := c.client.AddExternalContact(ctx, zoomphone.OptAddExternalContactReq{
		Value: zoomphone.AddExternalContactReq{
			Description:     util.ToPhoneOptString(dto.description),
			Email:           util.ToPhoneOptString(dto.email),
			ExtensionNumber: util.ToPhoneOptString(dto.extensionNumber),
			ID:              util.ToPhoneOptString(dto.id),
			Name:            dto.name.ValueString(),
			PhoneNumbers: lo.Map(dto.phoneNumbers, func(item types.String, index int) string {
				return item.ValueString()
			}),
			RoutingPath:      util.ToPhoneOptString(dto.routingPath),
			AutoCallRecorded: util.ToPhoneOptBool(dto.autoCallRecorded),
		},
		Set: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone external contact: %v", err)
	}

	return &createdDto{
		externalContactID: util.FromOptString(res.ExternalContactID),
	}, nil
}

func (c *crud) update(ctx context.Context, dto *updateDto) error {
	err := c.client.UpdateExternalContact(ctx, zoomphone.OptUpdateExternalContactReq{
		Value: zoomphone.UpdateExternalContactReq{
			Description:     util.ToPhoneOptString(dto.description),
			Email:           util.ToPhoneOptString(dto.email),
			ExtensionNumber: util.ToPhoneOptString(dto.extensionNumber),
			ID:              util.ToPhoneOptString(dto.id),
			Name:            util.ToPhoneOptString(dto.name),
			PhoneNumbers: lo.Map(dto.phoneNumbers, func(item types.String, index int) string {
				return item.ValueString()
			}),
			RoutingPath:      util.ToPhoneOptString(dto.routingPath),
			AutoCallRecorded: util.ToPhoneOptBool(dto.autoCallRecorded),
		},
		Set: true,
	}, zoomphone.UpdateExternalContactParams{
		ExternalContactId: dto.externalContactID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone external contact: %v", err)
	}

	return nil
}

func (c *crud) delete(ctx context.Context, externalContactId types.String) error {
	err := c.client.DeleteAExternalContact(ctx, zoomphone.DeleteAExternalContactParams{
		ExternalContactId: externalContactId.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone external contact: %v", err)
	}

	return nil
}
