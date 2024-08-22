package callqueuemember

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
	var callQueueMembers []*readDtoCallQueueMember
	nextPageToken := zoomphone.OptString{}
	for {
		ret, err := c.client.ListCallQueueMembers(ctx, zoomphone.ListCallQueueMembersParams{
			CallQueueId:   callQueueID.ValueString(),
			NextPageToken: nextPageToken,
			PageSize:      zoomphone.NewOptInt(300), // Constraints: Max 300
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
		callQueueMembers = append(callQueueMembers, lo.Map(ret.CallQueueMembers, func(member zoomphone.ListCallQueueMembersOKCallQueueMembersItem, _index int) *readDtoCallQueueMember {
			return &readDtoCallQueueMember{
				id:          util.FromOptString(member.ID),
				name:        util.FromOptString(member.Name),
				level:       util.FromOptString(member.Level),
				receiveCall: util.FromOptBool(member.ReceiveCall),
				extensionID: util.FromOptString(member.ExtensionID),
			}
		})...)
		if ret.NextPageToken.Value == "" {
			break
		}
		nextPageToken = ret.NextPageToken
	}

	return &readDto{
		callQueueMembers: callQueueMembers,
	}, nil
}

func (c *crud) readUsersByEmails(ctx context.Context, emails []types.String) (*readUsersDto, error) {
	return c.readUsersByCond(ctx, func(user zoomphone.ListPhoneUsersOKUsersItem) bool {
		return lo.ContainsBy(emails, func(email types.String) bool {
			return user.Email.Value == email.ValueString()
		})
	})
}

func (c *crud) readUsersByExtensionIDs(ctx context.Context, extensionIDs []types.String) (*readUsersDto, error) {
	return c.readUsersByCond(ctx, func(user zoomphone.ListPhoneUsersOKUsersItem) bool {
		return lo.ContainsBy(extensionIDs, func(extensionID types.String) bool {
			return user.ExtensionID.Value == extensionID.ValueString()
		})
	})
}

func (c *crud) readUsersByCond(ctx context.Context, cond func(u zoomphone.ListPhoneUsersOKUsersItem) bool) (*readUsersDto, error) {
	var users []*readUsersDtoUser
	nextPageToken := zoomphone.OptString{}
	for {
		res, err := c.client.ListPhoneUsers(ctx, zoomphone.ListPhoneUsersParams{
			NextPageToken: nextPageToken,
			PageSize:      zoomphone.NewOptInt(100), // Max 100
		})
		if err != nil {
			return nil, fmt.Errorf("error listing phone users: %v", err)
		}
		for _, user := range res.Users {
			isSearchedUser := cond(user)
			if isSearchedUser {
				users = append(users, &readUsersDtoUser{
					email:       util.FromOptString(user.Email),
					extensionID: util.FromOptString(user.ExtensionID),
				})
			}
		}
		if res.NextPageToken.Value == "" {
			break
		}
		nextPageToken = res.NextPageToken
	}
	return &readUsersDto{
		users: users,
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
			ID:    util.ToPhoneOptString(user.id),
			Email: util.ToPhoneOptString(user.email),
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
			return fmt.Errorf("error assigning phone call queue members: %v", err)
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
			return fmt.Errorf("error assigning phone call queue members: %v", err)
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
			return fmt.Errorf("error unassigning phone call queue members: %v", err)
		}
	}

	return nil
}

func (c *crud) unassignAll(ctx context.Context, callQueueID types.String) error {
	err := c.client.UnassignAllMembers(ctx, zoomphone.UnassignAllMembersParams{
		CallQueueId: callQueueID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 404 {
				return nil
			}
		}
		return fmt.Errorf("error unassigning all phone call queue members: %v", err)
	}
	return nil
}
