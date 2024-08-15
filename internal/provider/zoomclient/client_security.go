package zoomclient

import (
	"context"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomuser"
)

type ZoomPhoneClientSecurity struct {
	AccessToken string
}

func (c ZoomPhoneClientSecurity) OpenapiAuthorization(_ context.Context, _ string) (zoomphone.OpenapiAuthorization, error) {
	return zoomphone.OpenapiAuthorization{}, nil
}
func (c ZoomPhoneClientSecurity) OpenapiOAuth(_ context.Context, _ string) (zoomphone.OpenapiOAuth, error) {
	return zoomphone.OpenapiOAuth{Token: c.AccessToken}, nil
}

type ZoomUserClientSecurity struct {
	AccessToken string
}

func (c ZoomUserClientSecurity) OpenapiAuthorization(_ context.Context, _ string) (zoomuser.OpenapiAuthorization, error) {
	return zoomuser.OpenapiAuthorization{}, nil
}
func (c ZoomUserClientSecurity) OpenapiOAuth(_ context.Context, _ string) (zoomuser.OpenapiOAuth, error) {
	return zoomuser.OpenapiOAuth{Token: c.AccessToken}, nil
}
