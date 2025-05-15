package acceptance

import (
	"github.com/folio-sec/terraform-provider-zoom/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	ProviderConfig = `
provider "zoom" {
}`
)

var Provider = provider.New("test")().(*provider.ZoomProvider)
var TestAccProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"zoom": providerserver.NewProtocol6WithError(Provider),
}
