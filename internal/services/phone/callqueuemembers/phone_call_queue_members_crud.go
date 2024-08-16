package callqueuemembers

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
	ret, err := c.client.ListCallQueueMembers(ctx, zoomphone.ListCallQueueMembersParams{
		CallQueueId: callQueueID.ValueString(),
		// oops... zoom api spec doesn't allow page_token & page_size parameters
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone call queue members: %v", err)
	}

	callQueueMembers := lo.Map(ret.CallQueueMembers, func(member zoomphone.ListCallQueueMembersOKCallQueueMembersItem, _index int) *readDtoCallQueueMembers {
		return &readDtoCallQueueMembers{
			id:          util.FromOptString(member.ID),
			name:        util.FromOptString(member.Name),
			level:       util.FromOptString(member.Level),
			receiveCall: util.FromOptBool(member.ReceiveCall),
			extensionID: util.FromOptString(member.ExtensionID),
		}
	})
	return &readDto{
		callQueueMembers: callQueueMembers,
	}, nil
}

func (c *crud) assign(ctx context.Context, dto *assignDto) error {
	commonAreaIDs := make([]string, len(dto.commonAreaIDs))
	for i, id := range dto.commonAreaIDs {
		commonAreaIDs[i] = id.ValueString()
	}
	users := make([]zoomphone.AddMembersToCallQueueReqMembersUsersItem, len(dto.users))
	for i, user := range dto.users {
		users[i] = zoomphone.AddMembersToCallQueueReqMembersUsersItem{
			ID:    util.ToOptString(user.id),
			Email: util.ToOptString(user.email),
		}
	}

	// A maximum of 10 members can be added at a time.
	for _, commonAreaIDChunked := range lo.Chunk(commonAreaIDs, 10) {
		err := c.client.AddMembersToCallQueue(ctx, zoomphone.OptAddMembersToCallQueueReq{
			Value: zoomphone.AddMembersToCallQueueReq{
				Members: zoomphone.NewOptAddMembersToCallQueueReqMembers(zoomphone.AddMembersToCallQueueReqMembers{
					CommonAreaIds: commonAreaIDChunked,
				}),
			},
			Set: true,
		}, zoomphone.AddMembersToCallQueueParams{
			CallQueueId: dto.callQueueID.ValueString(),
		})
		if err != nil {
			return fmt.Errorf("error creating phone call queue members: %v", err)
		}
	}
	for _, userChunked := range lo.Chunk(users, 10) {
		err := c.client.AddMembersToCallQueue(ctx, zoomphone.OptAddMembersToCallQueueReq{
			Value: zoomphone.AddMembersToCallQueueReq{
				Members: zoomphone.NewOptAddMembersToCallQueueReqMembers(zoomphone.AddMembersToCallQueueReqMembers{
					Users: userChunked,
				}),
			},
			Set: true,
		}, zoomphone.AddMembersToCallQueueParams{
			CallQueueId: dto.callQueueID.ValueString(),
		})
		if err != nil {
			return fmt.Errorf("error creating phone call queue members: %v", err)
		}
	}

	return nil
}

func (c *crud) unassign(ctx context.Context, dto *unassignDto) error {
	for _, memberID := range dto.memberIDs {
		err := c.client.UnassignMemberFromCallQueue(ctx, zoomphone.UnassignMemberFromCallQueueParams{
			CallQueueId: dto.callQueueID.ValueString(),
			MemberId:    memberID.ValueString(),
		})
		if err != nil {
			var status *zoomphone.ErrorResponseStatusCode
			if errors.As(err, &status) {
				if status.StatusCode == 404 {
					continue
				}
			}
			return fmt.Errorf("error deleting phone call queue members: %v", err)
		}
	}

	return nil
}
