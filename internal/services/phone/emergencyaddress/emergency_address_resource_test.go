package emergencyaddress_test

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

func TestAccPhoneEmergencyAddress(t *testing.T) {
	td := acceptance.NewTestData(t, "zoom_phone_emergency_address", "test")

	emergencyAddressTestResource := newEmergencyAddressResource(td)
	updateEmergencyAddressTestResource := emergencyAddressTestResource.copyWithAddressLine1(fmt.Sprintf("acctest-%s", td.RandomStringOfLength(5)))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:             testAccPhoneEmergencyAddressResourceDestroy(td),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: acceptance.ProviderConfig + emergencyAddressTestResource.requiredConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(td.ResourceName, "address_line1", emergencyAddressTestResource.AddressLine1),
					resource.TestCheckResourceAttr(td.ResourceName, "city", emergencyAddressTestResource.City),
					resource.TestCheckResourceAttr(td.ResourceName, "country", emergencyAddressTestResource.Country),
					resource.TestCheckResourceAttr(td.ResourceName, "state", emergencyAddressTestResource.StateCode),
					resource.TestCheckResourceAttr(td.ResourceName, "zip", emergencyAddressTestResource.Zip),
					resource.TestCheckResourceAttrSet(td.ResourceName, "id"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "status"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "level"),
				),
			},
			// ImportState testing
			{
				ResourceName:      td.ResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: acceptance.ProviderConfig + updateEmergencyAddressTestResource.requiredConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(td.ResourceName, "address_line1", updateEmergencyAddressTestResource.AddressLine1),
					resource.TestCheckResourceAttr(td.ResourceName, "city", updateEmergencyAddressTestResource.City),
					resource.TestCheckResourceAttr(td.ResourceName, "country", updateEmergencyAddressTestResource.Country),
					resource.TestCheckResourceAttr(td.ResourceName, "state", updateEmergencyAddressTestResource.StateCode),
					resource.TestCheckResourceAttr(td.ResourceName, "zip", updateEmergencyAddressTestResource.Zip),
					resource.TestCheckResourceAttrSet(td.ResourceName, "id"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "status"),
					resource.TestCheckResourceAttrSet(td.ResourceName, "level"),
				),
			},
		},
	})
}

func testAccPhoneEmergencyAddressResourceDestroy(td acceptance.TestData) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for label, resourceState := range s.RootModule().Resources {
			if resourceState.Type != td.TerraformResourceType || label != td.ResourceName {
				continue
			}

			result, err := acceptance.Provider.ProviderData.PhoneClient.GetEmergencyAddress(context.Background(), zoomphone.GetEmergencyAddressParams{
				EmergencyAddressId: resourceState.Primary.ID,
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
				return fmt.Errorf("unable to read emergency address %q: %v", td.ResourceName, err)
			}
		}

		return nil
	}
}

func newEmergencyAddressResource(td acceptance.TestData) emergencyAddressResource {
	return emergencyAddressResource{
		ResourceLabel:         td.ResourceLabel,
		TerraformResourceType: td.TerraformResourceType,
		AddressLine1:          fmt.Sprintf("acctest-%s", td.RandomStringOfLength(5)),
		City:                  td.RandomStringOfLength(5),
		Country:               "JP",
		StateCode:             "Tokyo",
		Zip:                   "1000005",
	}
}
