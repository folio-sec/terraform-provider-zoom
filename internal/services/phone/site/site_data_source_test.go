package site_test

import (
	"fmt"
	"testing"

	"github.com/folio-sec/terraform-provider-zoom/internal/acceptance"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSite(t *testing.T) {
	td := acceptance.NewTestData(t, "zoom_phone_site", "test")

	siteTestResource := newSiteResource(td)
	dataConfig := fmt.Sprintf(`
data "zoom_phone_site" "test" {
  # Load data after the resource is created using the resource
  id = %[1]s
}
`,
		fmt.Sprintf("%s.id", td.ResourceName),
	)
	expectedDataSourceName := "data.zoom_phone_site.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:             testAccPhoneSiteResourceDestroy(td),
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: acceptance.ProviderConfig + siteTestResource.requiredConfig() + dataConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedDataSourceName, "name", siteTestResource.Name),
					resource.TestCheckResourceAttr(expectedDataSourceName, "main_auto_receptionist.name", siteTestResource.MainAutoReceptionistName),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "country.code"),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "country.name"),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "main_auto_receptionist.id"),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "sip_zone.id"),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "sip_zone.name"),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "level"),
				),
			},
		},
	})
}
