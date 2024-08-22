package autoreceptionist

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

func (c *crud) read(ctx context.Context, autoReceptionistID types.String) (*readDto, error) {
	detail, err := c.client.GetAutoReceptionistDetail(ctx, zoomphone.GetAutoReceptionistDetailParams{
		AutoReceptionistId: autoReceptionistID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone auto receptionist: %v", err)
	}

	return &readDto{
		autoReceptionistID:  autoReceptionistID,
		costCenter:          util.FromOptString(detail.CostCenter),
		department:          util.FromOptString(detail.Department),
		extensionID:         util.FromOptString(detail.ExtensionID),
		extensionNumber:     util.FromOptInt64(detail.ExtensionNumber),
		name:                util.FromOptString(detail.Name),
		timezone:            util.FromOptString(detail.Timezone),
		audioPromptLanguage: util.FromOptString(detail.AudioPromptLanguage),
	}, nil
}

func (c *crud) create(ctx context.Context, dto *createDto) (*createdDto, error) {
	res, err := c.client.AddAutoReceptionist(ctx, zoomphone.OptAddAutoReceptionistReq{
		Value: zoomphone.AddAutoReceptionistReq{
			Name: dto.name.ValueString(),
		},
		Set: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone auto receptionist: %v", err)
	}

	return &createdDto{
		autoReceptionistID: util.FromOptString(res.ID),
		name:               util.FromOptString(res.Name),
		extensionNumber:    util.FromOptInt64(res.ExtensionNumber),
	}, nil
}

func (c *crud) update(ctx context.Context, dto *updateDto) error {
	err := c.client.UpdateAutoReceptionist(ctx, zoomphone.OptUpdateAutoReceptionistReq{
		Value: zoomphone.UpdateAutoReceptionistReq{
			// CostCenter/Department: to remove it, need to pass empty string. not null.
			CostCenter:          zoomphone.NewOptString(util.ToPhoneOptString(dto.costCenter).Or("")),
			Department:          zoomphone.NewOptString(util.ToPhoneOptString(dto.department).Or("")),
			ExtensionNumber:     util.ToPhoneOptInt64(dto.extensionNumber),
			Name:                util.ToPhoneOptString(dto.name),
			AudioPromptLanguage: util.ToPhoneOptString(dto.audioPromptLanguage),
			Timezone:            util.ToPhoneOptString(dto.timezone),
		},
		Set: true,
	}, zoomphone.UpdateAutoReceptionistParams{
		AutoReceptionistId: dto.autoReceptionistID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone auto receptionist: %v", err)
	}

	return nil
}

func (c *crud) delete(ctx context.Context, autoReceptionistId types.String) error {
	err := c.client.DeleteAutoReceptionist(ctx, zoomphone.DeleteAutoReceptionistParams{
		AutoReceptionistId: autoReceptionistId.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 404 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone auto receptionist: %v", err)
	}

	return nil
}
