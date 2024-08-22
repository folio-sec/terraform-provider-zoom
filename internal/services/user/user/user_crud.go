package user

import (
	"context"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomuser"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

func newCrud(client *zoomuser.Client) *crud {
	return &crud{
		client: client,
	}
}

type crud struct {
	client *zoomuser.Client
}

func (c *crud) read(ctx context.Context, dto *readQueryDto) (*readDto, error) {
	var user []*readDtoUser
	nextPageToken := zoomuser.OptString{}
	for {
		ret, err := c.client.Users(ctx, zoomuser.UsersParams{
			PageSize:      zoomuser.NewOptInt(300), // max 300
			NextPageToken: nextPageToken,
			Status:        util.ToUserOptString(dto.status),
			RoleID:        util.ToUserOptString(dto.roleID),
			PageNumber:    util.ToUserOptString(dto.pageNumber),
			IncludeFields: util.ToUserOptString(dto.includeFields),
			License:       util.ToUserOptString(dto.license),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to read user: %v", err)
		}
		user = append(user, lo.Map(ret.Users, func(item zoomuser.UsersOKUsersItem, _index int) *readDtoUser {
			return &readDtoUser{
				id:               util.FromOptString(item.ID),
				email:            types.StringValue(item.Email),
				customAttributes: util.FromOptString(item.CustomAttributes),
				dept:             util.FromOptString(item.Dept),
				employeeUniqueID: util.FromOptString(item.EmployeeUniqueID),
				firstName:        util.FromOptString(item.FirstName),
				lastName:         util.FromOptString(item.LastName),
				groupIds: lo.Map(item.GroupIds, func(item string, index int) types.String {
					return types.StringValue(item)
				}),
				hostKey: util.FromOptString(item.HostKey),
				imGroupIds: lo.Map(item.ImGroupIds, func(item string, index int) types.String {
					return types.StringValue(item)
				}),
				planUnitedType: util.FromOptString(item.PlanUnitedType),
				pmi:            util.FromOptInt64(item.Pmi),
				roleID:         util.FromOptString(item.RoleID),
				status:         util.FromOptString(item.Status),
				typ:            types.Int32Value(int32(item.Type)),
				verified:       util.FromOptInt(item.Verified),
				displayName:    util.FromOptString(item.DisplayName),
			}
		})...)
		if ret.NextPageToken.Value == "" {
			break
		}
		nextPageToken = ret.NextPageToken
	}

	return &readDto{
		user: user,
	}, nil
}
