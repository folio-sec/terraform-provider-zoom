package site_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/acceptance"
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
