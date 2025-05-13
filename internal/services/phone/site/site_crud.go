package site

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

func (c *crud) read(ctx context.Context, siteID types.String) (*readDto, error) {
	detail, err := c.client.GetASite(ctx, zoomphone.GetASiteParams{
		SiteId: siteID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone site: %v", err)
	}

	return &readDto{
		id: util.FromOptString(detail.ID),
		country: readDtoCountry{
			code: util.FromOptString(detail.Country.Value.Code),
			name: util.FromOptString(detail.Country.Value.Name),
		},
		mainAutoReceptionist: readDtoMainAutoReceptionist{
			id:              util.FromOptString(detail.MainAutoReceptionist.Value.ID),
			name:            util.FromOptString(detail.MainAutoReceptionist.Value.Name),
			extensionID:     util.FromOptString(detail.MainAutoReceptionist.Value.ExtensionID),
			extensionNumber: util.FromOptInt64(detail.MainAutoReceptionist.Value.ExtensionNumber),
		},
		name:                 util.FromOptString(detail.Name),
		shortExtensionLength: util.FromOptInt(detail.ShortExtension.Value.Length),
		siteCode:             util.FromOptInt(detail.SiteCode),
		sipZone: readDtoSipZone{
			id:   util.FromOptString(detail.SipZone.Value.ID),
			name: util.FromOptString(detail.SipZone.Value.Name),
		},
		callerIDName:    util.FromOptString(detail.CallerIDName),
		level:           util.FromOptString(detail.Level),
		indiaStateCode:  util.FromOptString(detail.IndiaStateCode),
		indiaCity:       util.FromOptString(detail.IndiaCity),
		indiaSdcaNpa:    util.FromOptString(detail.IndiaSdcaNpa),
		indiaEntityName: util.FromOptString(detail.IndiaEntityName),
	}, nil
}

func (c *crud) create(ctx context.Context, dto *createDto) (*createdDto, error) {
	res, err := c.client.CreatePhoneSite(ctx, zoomphone.NewOptCreatePhoneSiteReq(
		zoomphone.CreatePhoneSiteReq{
			AutoReceptionistName:     dto.autoReceptionistName.ValueString(),
			SourceAutoReceptionistID: util.ToPhoneOptString(dto.sourceAutoReceptionistID),
			DefaultEmergencyAddress: zoomphone.CreatePhoneSiteReqDefaultEmergencyAddress{
				AddressLine1: dto.defaultEmergencyAddress.addressLine1.ValueString(),
				AddressLine2: util.ToPhoneOptString(dto.defaultEmergencyAddress.addressLine2),
				City:         dto.defaultEmergencyAddress.city.ValueString(),
				StateCode:    dto.defaultEmergencyAddress.stateCode.ValueString(),
				Country:      dto.defaultEmergencyAddress.countryCode.ValueString(),
				Zip:          dto.defaultEmergencyAddress.zip.ValueString(),
			},
			Name: dto.name.ValueString(),
			ShortExtension: zoomphone.NewOptCreatePhoneSiteReqShortExtension(
				zoomphone.CreatePhoneSiteReqShortExtension{
					Length: util.ToPhoneOptInt(dto.shortExtensionLength),
				},
			),
			SiteCode: util.ToPhoneOptInt(dto.siteCode),
			SipZone: zoomphone.NewOptCreatePhoneSiteReqSipZone(
				zoomphone.CreatePhoneSiteReqSipZone{
					ID: util.ToPhoneOptString(dto.sipZoneID),
				},
			),
			IndiaStateCode:  util.ToPhoneOptString(dto.indiaStateCode),
			IndiaCity:       util.ToPhoneOptString(dto.indiaCity),
			IndiaSdcaNpa:    util.ToPhoneOptString(dto.indiaSdcaNpa),
			IndiaEntityName: util.ToPhoneOptString(dto.indiaEntityName),
		},
	))
	if err != nil {
		return nil, fmt.Errorf("error creating phone site: %v", err)
	}

	return &createdDto{
		id:   util.FromOptString(res.ID),
		name: util.FromOptString(res.Name),
	}, nil
}

func (c *crud) update(ctx context.Context, dto *updateDto) error {
	err := c.client.UpdateSiteDetails(ctx,
		zoomphone.NewOptUpdateSiteDetailsReq(zoomphone.UpdateSiteDetailsReq{
			Name:     util.ToPhoneOptString(dto.name),
			SiteCode: util.ToPhoneOptInt(dto.siteCode),
			ShortExtension: lo.TernaryF(dto.shortExtension != nil, func() zoomphone.OptUpdateSiteDetailsReqShortExtension {
				return zoomphone.NewOptUpdateSiteDetailsReqShortExtension(zoomphone.UpdateSiteDetailsReqShortExtension{
					Length: util.ToPhoneOptInt(dto.shortExtension.length),
					Ranges: lo.Map(dto.shortExtension.ranges, func(item updateDtoShortExtensionRange, _ int) zoomphone.UpdateSiteDetailsReqShortExtensionRangesItem {
						return zoomphone.UpdateSiteDetailsReqShortExtensionRangesItem{
							RangeFrom: util.ToPhoneOptString(item.rangeFrom),
							RangeTo:   util.ToPhoneOptString(item.rangeTo),
						}
					}),
				})
			}, func() zoomphone.OptUpdateSiteDetailsReqShortExtension {
				return zoomphone.OptUpdateSiteDetailsReqShortExtension{}
			}),
			SipZone: zoomphone.NewOptUpdateSiteDetailsReqSipZone(zoomphone.UpdateSiteDetailsReqSipZone{
				ID: util.ToPhoneOptString(dto.sipZoneID),
			}),
			CallerIDName: util.ToPhoneOptString(dto.callerIDName),
		}),
		zoomphone.UpdateSiteDetailsParams{
			SiteId: dto.id.ValueString(),
		})
	if err != nil {
		return fmt.Errorf("error updating phone site: %v", err)
	}

	return nil
}

func (c *crud) delete(ctx context.Context, siteID types.String, transferSiteID types.String) error {
	err := c.client.DeletePhoneSite(ctx, zoomphone.DeletePhoneSiteParams{
		SiteId:         siteID.ValueString(),
		TransferSiteID: transferSiteID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 404 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone site: %v", err)
	}

	return nil
}

func (c *crud) readMain(ctx context.Context) (*readDto, error) {
	nextPageToken := zoomphone.OptString{}
	for {
		ret, err := c.client.ListPhoneSites(ctx, zoomphone.ListPhoneSitesParams{
			NextPageToken: nextPageToken,
			PageSize:      zoomphone.NewOptInt(300), // max 300
		})
		if err != nil {
			return nil, fmt.Errorf("unable to read sites: %v", err)
		}

		for _, site := range ret.Sites {
			if site.Level.IsSet() && site.Level.Value == "main" {
				return c.read(ctx, util.FromOptString(site.ID))
			}
		}

		if ret.NextPageToken.Value == "" {
			break
		}
		nextPageToken = ret.NextPageToken
	}

	return nil, fmt.Errorf("main site not found")
}
