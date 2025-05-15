package site_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/acceptance"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type siteResource struct {
	ResourceLabel            string
	TerraformResourceType    string
	Name                     string
	MainAutoReceptionistName string
	AddressLine1             string
	City                     string
	Country                  string
	StateCode                string
	Zip                      string
	SiteCode                 int
}

func (sr siteResource) copyWithName(value string) siteResource {
	return siteResource{
		ResourceLabel:            sr.ResourceLabel,
		TerraformResourceType:    sr.TerraformResourceType,
		Name:                     value,
		MainAutoReceptionistName: sr.MainAutoReceptionistName,
		AddressLine1:             sr.AddressLine1,
		City:                     sr.City,
		Country:                  sr.Country,
		StateCode:                sr.StateCode,
		Zip:                      sr.Zip,
		SiteCode:                 sr.SiteCode,
	}
}

func (sr siteResource) requiredConfig() string {
	return fmt.Sprintf(`
resource %[1]q %[2]q {
  name = %[3]q
  main_auto_receptionist = {
    name = %[4]q
  }
  default_emergency_address = {
    address_line1 = %[5]q
    city = %[6]q
    country = %[7]q
    state_code = %[8]q
    zip = %[9]q
  }
}
`,
		sr.TerraformResourceType,
		sr.ResourceLabel,
		sr.Name,
		sr.MainAutoReceptionistName,
		sr.AddressLine1,
		sr.City,
		sr.Country,
		sr.StateCode,
		sr.Zip,
	)
}

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

func testAccPhoneSiteResourceDestroy(td acceptance.TestData) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for label, resourceState := range s.RootModule().Resources {
			if resourceState.Type != td.TerraformResourceType || label != td.ResourceName {
				continue
			}

			result, err := acceptance.Provider.ProviderData.PhoneClient.GetASite(context.Background(), zoomphone.GetASiteParams{
				SiteId: resourceState.Primary.ID,
			})
			if result == nil && err == nil {
				return fmt.Errorf("should have either an error or a result when checking if %q has been destroyed", td.ResourceName)
			}
			if result != nil {
				return fmt.Errorf("%q still exists", td.ResourceName)
			}
			if err != nil {
				var status *zoomphone.ErrorResponseStatusCode
				if errors.As(err, &status) {
					if status.StatusCode == 400 && status.Response.Code.Value == 300 {
						// already deleted
						return nil
					}
				}
				return fmt.Errorf("unable to read phone site %q: %v", td.ResourceName, err)
			}
		}

		return nil
	}
}

func newSiteResource(td acceptance.TestData) siteResource {
	return siteResource{
		ResourceLabel:            td.ResourceLabel,
		TerraformResourceType:    td.TerraformResourceType,
		Name:                     fmt.Sprintf("acctest-%s", td.RandomStringOfLength(5)),
		MainAutoReceptionistName: fmt.Sprintf("acctest-%s", td.RandomStringOfLength(5)),
		AddressLine1:             td.RandomStringOfLength(5),
		City:                     td.RandomStringOfLength(5),
		Country:                  "JP",
		StateCode:                "Tokyo",
		Zip:                      "1000005",
		SiteCode:                 10,
	}
}
