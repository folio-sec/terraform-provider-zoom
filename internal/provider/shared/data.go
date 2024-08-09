package shared

import "github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"

// ProviderData is the data that is passed to objects for provider.
type ProviderData struct {
	PhoneMasterClient *zoomphone.Client
}
