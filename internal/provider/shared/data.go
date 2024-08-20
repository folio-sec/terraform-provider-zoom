package shared

import (
	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomuser"
)

// ProviderData is the data that is passed to objects for provider.
type ProviderData struct {
	PhoneClient *zoomphone.Client
	UserClient  *zoomuser.Client
}
