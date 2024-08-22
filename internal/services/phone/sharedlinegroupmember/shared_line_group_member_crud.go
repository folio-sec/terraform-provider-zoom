package sharedlinegroupgroupmembers

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
	var commonAreas []*readDtoCommonArea
	var users []*readDtoUser
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
		return nil, fmt.Errorf("unable to read phone shared line group members: %v", err)
	}
	if ret.Members.Set {
		commonAreas = lo.Map(ret.Members.Value.CommonAreas, func(member zoomphone.GetASharedLineGroupOKMembersCommonAreasItem, _index int) *readDtoCommonArea {
			return &readDtoCommonArea{
				id:          util.FromOptString(member.ID),
				name:        util.FromOptString(member.Name),
				extensionID: util.FromOptString(member.ExtensionID),
			}
		})
		users = lo.Map(ret.Members.Value.Users, func(member zoomphone.GetASharedLineGroupOKMembersUsersItem, _index int) *readDtoUser {
			return &readDtoUser{
				id:          util.FromOptString(member.ID),
				name:        util.FromOptString(member.Name),
				extensionID: util.FromOptString(member.ExtensionID),
			}
		})
	}

	return &readDto{
		commonAreas: commonAreas,
		users:       users,
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
	users := make([]zoomphone.AddMembersToSharedLineGroupReqMembersUsersItem, len(dto.users))
	for i, user := range dto.users {
		users[i] = zoomphone.AddMembersToSharedLineGroupReqMembersUsersItem{
			ID:    util.ToPhoneOptString(user.id),
			Email: util.ToPhoneOptString(user.email),
		}
	}

	// A maximum of 10 members can be added at a time.
	for _, commonAreaIDChunked := range lo.Chunk(commonAreaIDs, 10) {
		err := c.client.AddMembersToSharedLineGroup(ctx, zoomphone.OptAddMembersToSharedLineGroupReq{
			Value: zoomphone.AddMembersToSharedLineGroupReq{
				Members: zoomphone.NewOptAddMembersToSharedLineGroupReqMembers(zoomphone.AddMembersToSharedLineGroupReqMembers{
					CommonAreaIds: commonAreaIDChunked,
				}),
			},
			Set: true,
		}, zoomphone.AddMembersToSharedLineGroupParams{
			SharedLineGroupId: dto.sharedLineGroupID.ValueString(),
		})
		if err != nil {
			return fmt.Errorf("error assigning phone shared line group members: %v", err)
		}
	}
	for _, userChunked := range lo.Chunk(users, 10) {
		err := c.client.AddMembersToSharedLineGroup(ctx, zoomphone.OptAddMembersToSharedLineGroupReq{
			Value: zoomphone.AddMembersToSharedLineGroupReq{
				Members: zoomphone.NewOptAddMembersToSharedLineGroupReqMembers(zoomphone.AddMembersToSharedLineGroupReqMembers{
					Users: userChunked,
				}),
			},
			Set: true,
		}, zoomphone.AddMembersToSharedLineGroupParams{
			SharedLineGroupId: dto.sharedLineGroupID.ValueString(),
		})
		if err != nil {
			return fmt.Errorf("error assigning phone shared line group members: %v", err)
		}
	}

	return nil
}

func (c *crud) unassign(ctx context.Context, dto *unassignDto) error {
	for _, memberID := range dto.memberIDs {
		err := c.client.DeleteAMemberSLG(ctx, zoomphone.DeleteAMemberSLGParams{
			SharedLineGroupId: dto.sharedLineGroupID.ValueString(),
			MemberId:          memberID.ValueString(),
		})
		if err != nil {
			var status *zoomphone.ErrorResponseStatusCode
			if errors.As(err, &status) {
				if status.StatusCode == 404 {
					continue
				}
			}
			return fmt.Errorf("error unassigning phone shared line group members: %v", err)
		}
	}

	return nil
}

func (c *crud) unassignAll(ctx context.Context, sharedLineGroupID types.String) error {
	err := c.client.DeleteMembersOfSLG(ctx, zoomphone.DeleteMembersOfSLGParams{
		SharedLineGroupId: sharedLineGroupID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 404 {
				return nil
			}
		}
		return fmt.Errorf("error unassigning all phone shared line group members: %v", err)
	}
	return nil
}
