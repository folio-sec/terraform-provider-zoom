package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
)

func newUserCrud(client *zoomphone.Client) *userCrud {
	return &userCrud{
		client: client,
	}
}

type userCrud struct {
	client *zoomphone.Client
}

func (c *userCrud) Read(ctx context.Context, zoomUserID string) (*readDto, error) {
	detail, err := c.client.PhoneUser(ctx, zoomphone.PhoneUserParams{
		UserId: zoomUserID,
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone user: %v", err)
	}

	return &readDto{}, nil
}

func (c *userCrud) Create(ctx context.Context, dto PhoneAutoReceptionistCreateDto) (*PhoneAutoReceptionistCreatedDto, error) {
	res, err := c.client.AddAutoReceptionist(ctx, zoomphone.OptAddAutoReceptionistReq{
		Value: zoomphone.AddAutoReceptionistReq{
			Name: dto.Name.ValueString(),
		},
		Set: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone auto receptionist: %v", err)
	}
	if bad, ok := res.(*zoomphone.AddAutoReceptionistBadRequest); ok {
		return nil, fmt.Errorf("error creating phone auto receptionist: bad request %v", bad)
	}
	if created, ok := res.(*zoomphone.AddAutoReceptionistCreated); ok {
		return &PhoneAutoReceptionistCreatedDto{
			ID:              util.FromOptString(created.ID),
			Name:            util.FromOptString(created.Name),
			ExtensionNumber: util.FromOptInt64(created.ExtensionNumber),
		}, nil
	}
	return nil, fmt.Errorf("error creating phone auto receptionist: invalid implementation %v", res)
}

func (c *userCrud) Update(ctx context.Context, dto PhoneAutoReceptionistUpdateDto) error {
	_, err := c.client.UpdateAutoReceptionist(ctx, zoomphone.OptUpdateAutoReceptionistReq{
		Value: zoomphone.UpdateAutoReceptionistReq{
			CostCenter:          util.ToOptString(dto.CostCenter),
			Department:          util.ToOptString(dto.Department),
			ExtensionNumber:     util.ToOptInt64(dto.ExtensionNumber),
			Name:                util.ToOptString(dto.Name),
			AudioPromptLanguage: util.ToOptString(dto.AudioPromptLanguage),
			Timezone:            util.ToOptString(dto.Timezone),
		},
		Set: true,
	}, zoomphone.UpdateAutoReceptionistParams{
		AutoReceptionistId: dto.AutoReceptionistID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone auto receptionist: %v", err)
	}
	return nil
}

func (c *userCrud) Delete(ctx context.Context, autoReceptionistId string) error {
	_, err := c.client.DeleteAutoReceptionist(ctx, zoomphone.DeleteAutoReceptionistParams{
		AutoReceptionistId: autoReceptionistId,
	})
	if err != nil {
		return fmt.Errorf("error deleting phone auto receptionist: %v", err)
	}
	return nil
}
