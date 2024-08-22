package phonenumbers

import (
	"context"
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

func (c *crud) read(ctx context.Context, dto *readQueryDto) (*readDto, error) {
	var phoneNumbers []*readDtoPhoneNumber
	nextPageToken := zoomphone.OptString{}
	for {
		ret, err := c.client.ListAccountPhoneNumbers(ctx, zoomphone.ListAccountPhoneNumbersParams{
			NextPageToken:  nextPageToken,
			PageSize:       zoomphone.NewOptInt(300), // max 300
			Type:           util.ToPhoneOptString(dto.typ),
			ExtensionType:  util.ToPhoneOptString(dto.extensionType),
			NumberType:     util.ToPhoneOptString(dto.numberType),
			PendingNumbers: util.ToPhoneOptBool(dto.pendingNumbers),
			SiteID:         util.ToPhoneOptString(dto.siteID),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to read phone numbers: %v", err)
		}
		phoneNumbers = append(phoneNumbers, lo.Map(ret.PhoneNumbers, func(item zoomphone.ListAccountPhoneNumbersOKPhoneNumbersItem, _index int) *readDtoPhoneNumber {
			capability := lo.Map(item.Capability, func(item string, index int) types.String {
				return types.StringValue(item)
			})
			var assignee *readDtoPhoneNumberAssignee
			if item.Assignee.IsSet() {
				assignee = &readDtoPhoneNumberAssignee{
					extensionNumber: util.FromOptInt64(item.Assignee.Value.ExtensionNumber),
					id:              util.FromOptString(item.Assignee.Value.ID),
					name:            util.FromOptString(item.Assignee.Value.Name),
					typ:             util.FromOptString(item.Assignee.Value.Type),
				}
			}
			var carrier *readDtoPhoneNumberCarrier
			if item.Carrier.IsSet() {
				carrier = &readDtoPhoneNumberCarrier{
					code: util.FromOptInt(item.Carrier.Value.Code),
					name: util.FromOptString(item.Carrier.Value.Name),
				}
			}
			var emergencyAddress *readDtoPhoneNumberEmergencyAddress
			if item.EmergencyAddress.IsSet() {
				emergencyAddress = &readDtoPhoneNumberEmergencyAddress{
					addressLine1: util.FromOptString(item.EmergencyAddress.Value.AddressLine1),
					addressLine2: util.FromOptString(item.EmergencyAddress.Value.AddressLine2),
					city:         util.FromOptString(item.EmergencyAddress.Value.City),
					country:      util.FromOptString(item.EmergencyAddress.Value.Country),
					stateCode:    util.FromOptString(item.EmergencyAddress.Value.StateCode),
					zip:          util.FromOptString(item.EmergencyAddress.Value.Zip),
				}
			}
			var sipGroup *readDtoPhoneNumberSipGroup
			if item.SipGroup.IsSet() {
				sipGroup = &readDtoPhoneNumberSipGroup{
					displayName: util.FromOptString(item.SipGroup.Value.DisplayName),
					id:          util.FromOptString(item.SipGroup.Value.ID),
				}
			}
			var site *readDtoPhoneNumberSite
			if item.Site.IsSet() {
				site = &readDtoPhoneNumberSite{
					id:   util.FromOptString(item.Site.Value.ID),
					name: util.FromOptString(item.Site.Value.Name),
				}
			}
			return &readDtoPhoneNumber{
				assignee:                   assignee,
				capability:                 capability,
				carrier:                    carrier,
				displayName:                util.FromOptString(item.DisplayName),
				emergencyAddress:           emergencyAddress,
				emergencyAddressStatus:     util.FromOptInt(item.EmergencyAddressStatus),
				emergencyAddressUpdateTime: util.FromOptString(item.EmergencyAddressUpdateTime),
				id:                         util.FromOptString(item.ID),
				location:                   util.FromOptString(item.Location),
				number:                     util.FromOptString(item.Number),
				numberType:                 util.FromOptString(item.NumberType),
				sipGroup:                   sipGroup,
				site:                       site,
				source:                     util.FromOptString(item.Source),
				status:                     util.FromOptString(item.Status),
			}
		})...)
		if ret.NextPageToken.Value == "" {
			break
		}
		nextPageToken = ret.NextPageToken
	}

	return &readDto{
		phoneNumbers: phoneNumbers,
	}, nil
}
