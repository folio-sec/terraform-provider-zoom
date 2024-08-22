package blockedlist

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

func (c *crud) read(ctx context.Context, blockedListID types.String) (*readDto, error) {
	detail, err := c.client.GetABlockedList(ctx, zoomphone.GetABlockedListParams{
		BlockedListId: blockedListID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone blocked list: %v", err)
	}

	return &readDto{
		blockedListID: blockedListID,
		blockType:     util.FromOptString(detail.BlockType),
		comment:       util.FromOptString(detail.Comment),
		matchType:     util.FromOptString(detail.MatchType),
		phoneNumber:   util.FromOptString(detail.PhoneNumber),
		status:        util.FromOptString(detail.Status),
	}, nil
}

func (c *crud) create(ctx context.Context, dto *createDto) (*createdDto, error) {
	res, err := c.client.AddAnumberToBlockedList(ctx, zoomphone.OptAddAnumberToBlockedListReq{
		Value: zoomphone.AddAnumberToBlockedListReq{
			BlockType:   util.ToPhoneOptString(dto.blockType),
			Comment:     util.ToPhoneOptString(dto.comment),
			MatchType:   util.ToPhoneOptString(dto.matchType),
			PhoneNumber: util.ToPhoneOptString(dto.phoneNumber),
			Status:      util.ToPhoneOptString(dto.status),
		},
		Set: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone blocked list: %v", err)
	}

	return &createdDto{
		blockedListID: util.FromOptString(res.ID),
	}, nil
}

func (c *crud) delete(ctx context.Context, blockedListId types.String) error {
	err := c.client.DeleteABlockedList(ctx, zoomphone.DeleteABlockedListParams{
		BlockedListId: blockedListId.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 404 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone blocked list: %v", err)
	}

	return nil
}
