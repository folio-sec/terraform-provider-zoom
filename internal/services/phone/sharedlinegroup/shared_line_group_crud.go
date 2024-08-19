package sharedlinegroupgroup

import (
	"context"
	"errors"
	"fmt"

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

func (c *crud) read(ctx context.Context, sharedLineGroupID types.String) (*readDto, error) {
	detail, err := c.client.GetASharedLineGroup(ctx, zoomphone.GetASharedLineGroupParams{
		SharedLineGroupId: sharedLineGroupID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone shared line group: %v", err)
	}

	var site *readDtoSite
	if detail.Site.IsSet() {
		site = &readDtoSite{
			id:   util.FromOptString(detail.Site.Value.ID),
			name: util.FromOptString(detail.Site.Value.Name),
		}
	}
	return &readDto{
		sharedLineGroupID: sharedLineGroupID,
		displayName:       util.FromOptString(detail.DisplayName),
		extensionID:       util.FromOptString(detail.ExtensionID),
		extensionNumber:   util.FromOptInt64(detail.ExtensionNumber),
		primaryNumber:     util.FromOptString(detail.PrimaryNumber),
		status:            util.FromOptString(detail.Status),
		site:              site,
	}, nil
}

func (c *crud) create(ctx context.Context, dto *createDto) (*createdDto, error) {
	res, err := c.client.CreateASharedLineGroup(ctx, zoomphone.OptCreateASharedLineGroupReq{
		Value: zoomphone.CreateASharedLineGroupReq{
			DisplayName:     dto.displayName.ValueString(),
			Description:     util.ToOptString(dto.description),
			ExtensionNumber: util.ToOptInt64(dto.extensionNumber),
			SiteID:          util.ToOptString(dto.siteID),
		},
		Set: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone shared line group: %v", err)
	}

	return &createdDto{
		sharedLineGroupID: util.FromOptString(res.ID),
	}, nil
}

func (c *crud) update(ctx context.Context, dto *updateDto) error {
	err := c.client.UpdateASharedLineGroup(ctx, zoomphone.OptUpdateASharedLineGroupReq{
		Value: zoomphone.UpdateASharedLineGroupReq{
			ExtensionNumber: util.ToOptInt64(dto.extensionNumber),
			DisplayName:     util.ToOptString(dto.displayName),
			Status:          util.ToOptString(dto.status),
			// PrimeNumber is managed by "shared_line_group_phone_numbers" resource
			// PrimaryNumber: util.ToOptString(dto.primeNumber),
		},
		Set: true,
	}, zoomphone.UpdateASharedLineGroupParams{
		SharedLineGroupId: dto.sharedLineGroupID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone shared line group: %v", err)
	}

	return nil
}

func (c *crud) delete(ctx context.Context, sharedLineGroupId types.String) error {
	err := c.client.DeleteASharedLineGroup(ctx, zoomphone.DeleteASharedLineGroupParams{
		SharedLineGroupId: sharedLineGroupId.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 404 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone shared line group: %v", err)
	}

	return nil
}
