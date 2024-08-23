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

func (c *crud) list(ctx context.Context, dto listQueryDto) (*listDto, error) {
	var users []listDtoUser
	nextPageToken := zoomuser.OptString{}

	for {
		ret, err := c.client.Users(ctx, zoomuser.UsersParams{
			Status:        util.ToUserOptString(dto.status),
			RoleID:        util.ToUserOptString(dto.roleID),
			IncludeFields: util.ToUserOptString(dto.includeFields),
			License:       util.ToUserOptString(dto.license),
			PageSize:      zoomuser.NewOptInt(300),
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to read users: %v", err)
		}

		users = append(users, lo.Map(ret.Users, func(item zoomuser.UsersOKUsersItem, _ int) listDtoUser {
			return listDtoUser{
				userID: util.FromOptString(item.ID),
				customAttributes: lo.Map(item.CustomAttributes, func(item zoomuser.UsersOKUsersItemCustomAttributesItem, _ int) listDtoUserCustomAttribute {
					return listDtoUserCustomAttribute{
						key:   util.FromOptString(item.Key),
						name:  util.FromOptString(item.Name),
						value: util.FromOptString(item.Value),
					}
				}),
				dept:             util.FromOptString(item.Dept),
				displayName:      util.FromOptString(item.DisplayName),
				email:            types.StringValue(item.Email),
				employeeUniqueID: util.FromOptString(item.EmployeeUniqueID),
				firstName:        util.FromOptString(item.FirstName),
				groupIDs: lo.Map(item.GroupIds, func(groupID string, _ int) types.String {
					return types.StringValue(groupID)
				}),
				hostKey: util.FromOptString(item.HostKey),
				imGroupIDs: lo.Map(item.ImGroupIds, func(imGroupID string, _ int) types.String {
					return types.StringValue(imGroupID)
				}),
				lastClientVersion: util.FromOptString(item.LastClientVersion),
				lastLoginTime:     util.FromOptDateTime(item.LastLoginTime),
				lastName:          util.FromOptString(item.LastName),
				planUnitedType:    util.FromOptString(item.PlanUnitedType),
				pmi:               util.FromOptInt64(item.Pmi),
				roleID:            util.FromOptString(item.RoleID),
				status:            util.FromOptString(item.Status),
				timezone:          util.FromOptString(item.Timezone),
				userType:          types.Int32Value(int32(item.Type)),
				userCreatedAt:     util.FromOptDateTime(item.UserCreatedAt),
				verified:          util.FromOptInt(item.Verified),
			}
		})...)

		if !ret.NextPageToken.IsSet() || ret.NextPageToken.Value == "" {
			break
		}
		nextPageToken = ret.NextPageToken
	}

	return &listDto{
		users: users,
	}, nil
}
