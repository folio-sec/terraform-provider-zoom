package emergencyaddress

import (
	"context"
	"errors"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type crud struct {
	client *zoomphone.Client
}

func newCrud(client *zoomphone.Client) *crud {
	return &crud{
		client: client,
	}
}

func (c *crud) create(ctx context.Context, req *createDto) (*readDto, error) {
	res, err := c.client.AddEmergencyAddress(ctx, zoomphone.NewOptAddEmergencyAddressReq(
		zoomphone.AddEmergencyAddressReq{
			AddressLine1: req.addressLine1.ValueString(),
			AddressLine2: util.ToPhoneOptString(req.addressLine2),
			City:         req.city.ValueString(),
			Country:      req.country.ValueString(),
			IsDefault:    util.ToPhoneOptBool(req.isDefault),
			SiteID:       util.ToPhoneOptString(req.siteID),
			StateCode:    req.state.ValueString(),
			Zip:          req.zip.ValueString(),
		}))
	if err != nil {
		return nil, err
	}

	return &readDto{
		id:           util.FromOptString(res.ID),
		addressLine1: util.FromOptString(res.AddressLine1),
		addressLine2: util.FromOptString(res.AddressLine2),
		city:         util.FromOptString(res.City),
		country:      util.FromOptString(res.Country),
		isDefault:    util.FromOptBool(res.IsDefault),
		level:        util.FromOptInt(res.Level),
		site: readDtoSite{
			ID:   util.FromOptString(res.Site.Value.ID),
			Name: util.FromOptString(res.Site.Value.Name),
		},
		owner: readDtoOwner{
			ID:              util.FromOptString(res.Owner.Value.ID),
			ExtensionNumber: util.FromOptInt64(res.Owner.Value.ExtensionNumber),
			Name:            util.FromOptString(res.Owner.Value.Name),
		},
		stateCode: util.FromOptString(res.StateCode),
		status:    util.FromOptInt(res.Status),
		zip:       util.FromOptString(res.Zip),
	}, nil
}

func (c *crud) read(ctx context.Context, emergencyAddressID types.String) (*readDto, error) {
	res, err := c.client.GetEmergencyAddress(ctx, zoomphone.GetEmergencyAddressParams{
		EmergencyAddressId: emergencyAddressID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone emergency address: %v", err)
	}

	return &readDto{
		id:           util.FromOptString(res.ID),
		addressLine1: util.FromOptString(res.AddressLine1),
		addressLine2: util.FromOptString(res.AddressLine2),
		city:         util.FromOptString(res.City),
		country:      util.FromOptString(res.Country),
		isDefault:    util.FromOptBool(res.IsDefault),
		level:        util.FromOptInt(res.Level),
		site: readDtoSite{
			ID:   util.FromOptString(res.Site.Value.ID),
			Name: util.FromOptString(res.Site.Value.Name),
		},
		owner: readDtoOwner{
			ID:              util.FromOptString(res.Owner.Value.ID),
			ExtensionNumber: util.FromOptInt64(res.Owner.Value.ExtensionNumber),
			Name:            util.FromOptString(res.Owner.Value.Name),
		},
		stateCode: util.FromOptString(res.StateCode),
		status:    util.FromOptInt(res.Status),
		zip:       util.FromOptString(res.Zip),
	}, nil
}

func (c *crud) update(ctx context.Context, req *updateDto) error {
	_, err := c.client.UpdateEmergencyAddress(ctx, zoomphone.NewOptUpdateEmergencyAddressReq(
		zoomphone.UpdateEmergencyAddressReq{
			AddressLine1: zoomphone.NewOptString(req.addressLine1.ValueString()),
			AddressLine2: zoomphone.NewOptString(req.addressLine2.ValueString()),
			City:         zoomphone.NewOptString(req.city.ValueString()),
			Country:      zoomphone.NewOptString(req.country.ValueString()),
			IsDefault:    zoomphone.NewOptBool(req.isDefault.ValueBool()),
			StateCode:    zoomphone.NewOptString(req.state.ValueString()),
			Zip:          zoomphone.NewOptString(req.zip.ValueString()),
		}), zoomphone.UpdateEmergencyAddressParams{
		EmergencyAddressId: req.id.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone emergency address: %v", err)
	}

	return nil
}

func (c *crud) delete(ctx context.Context, emergencyAddressID types.String) error {
	err := c.client.DeleteEmergencyAddress(ctx, zoomphone.DeleteEmergencyAddressParams{
		EmergencyAddressId: emergencyAddressID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 404 {
				return nil // already deleted
			}
		}
		return fmt.Errorf("error deleting phone emergency address: %v", err)
	}

	return nil
}
