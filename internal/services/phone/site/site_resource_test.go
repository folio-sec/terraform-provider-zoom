package site_test

import (
	"fmt"
	"testing"

	"github.com/folio-sec/terraform-provider-zoom/internal/acceptance"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPhoneSite(t *testing.T) {
	td := acceptance.NewTestData(t, "zoom_phone_site", "test")

	siteTestResource := newSiteResource(td)
	updatesiteTestResource := siteTestResource.copyWithName(fmt.Sprintf("acctest-%s", td.RandomStringOfLength(5)))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:             testAccPhoneSiteResourceDestroy(td),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: acceptance.ProviderConfig + siteTestResource.requiredConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(td.ResourceName, "name", siteTestResource.Name),
					resource.TestCheckResourceAttr(td.ResourceName, "main_auto_receptionist.name", siteTestResource.MainAutoReceptionistName),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.address_line1", siteTestResource.AddressLine1),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.city", siteTestResource.City),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.country", siteTestResource.Country),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.state_code", siteTestResource.StateCode),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.zip", siteTestResource.Zip),
					resource.TestCheckResourceAttrSet(td.ResourceName, "id"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "main_auto_receptionist.id"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "short_extension.length"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "sip_zone_id"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "level"),
				),
			},
			// ImportState testing
			{
				ResourceName:            td.ResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_emergency_address"},
			},
			// Update and Read testing
			{
				Config: acceptance.ProviderConfig + updatesiteTestResource.requiredConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(td.ResourceName, "name", updatesiteTestResource.Name),
					resource.TestCheckResourceAttr(td.ResourceName, "main_auto_receptionist.name", updatesiteTestResource.MainAutoReceptionistName),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.address_line1", updatesiteTestResource.AddressLine1),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.city", updatesiteTestResource.City),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.country", updatesiteTestResource.Country),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.state_code", updatesiteTestResource.StateCode),
					resource.TestCheckResourceAttr(td.ResourceName, "default_emergency_address.zip", updatesiteTestResource.Zip),
					resource.TestCheckResourceAttrSet(td.ResourceName, "id"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "main_auto_receptionist.id"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "short_extension.length"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "sip_zone_id"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "level"),
				),
			},
		},
	})
}
