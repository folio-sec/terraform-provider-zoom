package callqueue

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

func (c *crud) read(ctx context.Context, callQueueID types.String) (*readDto, error) {
	detail, err := c.client.GetACallQueue(ctx, zoomphone.GetACallQueueParams{
		CallQueueId: callQueueID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone call queue: %v", err)
	}

	var site *readDtoSite
	if detail.Site.IsSet() {
		site = &readDtoSite{
			id:   util.FromOptString(detail.Site.Value.ID),
			name: util.FromOptString(detail.Site.Value.Name),
		}
	}
	return &readDto{
		callQueueID:     callQueueID,
		costCenter:      util.FromOptString(detail.CostCenter),
		department:      util.FromOptString(detail.Department),
		extensionID:     util.FromOptString(detail.ExtensionID),
		extensionNumber: util.FromOptInt64(detail.ExtensionNumber),
		name:            util.FromOptString(detail.Name),
		// description: util.FromOptString(detail.Description), // api doesn't support description
		site:   site,
		status: util.FromOptString(detail.Status),
	}, nil
}

func (c *crud) create(ctx context.Context, dto *createDto) (*createdDto, error) {
	res, err := c.client.CreateCallQueue(ctx, zoomphone.OptCreateCallQueueReq{
		Value: zoomphone.CreateCallQueueReq{
			Name: dto.name.ValueString(),
			// CostCenter/Department: to remove it, need to pass empty string. not null.
			CostCenter:      zoomphone.NewOptString(util.ToOptString(dto.costCenter).Or("")),
			Department:      zoomphone.NewOptString(util.ToOptString(dto.department).Or("")),
			Description:     util.ToOptString(dto.description),
			ExtensionNumber: util.ToOptInt64(dto.extensionNumber),
			SiteID:          util.ToOptString(dto.siteID),
		},
		Set: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone call queue: %v", err)
	}

	return &createdDto{
		callQueueID: util.FromOptString(res.ID),
	}, nil
}

func (c *crud) update(ctx context.Context, dto *updateDto) error {
	err := c.client.UpdateCallQueue(ctx, zoomphone.OptUpdateCallQueueReq{
		Value: zoomphone.UpdateCallQueueReq{
			// CostCenter/Department: to remove it, need to pass empty string. not null.
			CostCenter:      zoomphone.NewOptString(util.ToOptString(dto.costCenter).Or("")),
			Department:      zoomphone.NewOptString(util.ToOptString(dto.department).Or("")),
			Description:     util.ToOptString(dto.description),
			ExtensionNumber: util.ToOptInt64(dto.extensionNumber),
			Name:            util.ToOptString(dto.name),
			Status:          util.ToOptString(dto.status),
			SiteID:          util.ToOptString(dto.siteID),
		},
		Set: true,
	}, zoomphone.UpdateCallQueueParams{
		CallQueueId: dto.callQueueID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone call queue: %v", err)
	}

	return nil
}

func (c *crud) delete(ctx context.Context, callQueueId types.String) error {
	err := c.client.DeleteACallQueue(ctx, zoomphone.DeleteACallQueueParams{
		CallQueueId: callQueueId.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 404 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone call queue: %v", err)
	}

	return nil
}
