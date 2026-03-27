package emergencyaddress_test

import (
	"fmt"
	"testing"

	"github.com/folio-sec/terraform-provider-zoom/internal/acceptance"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceEmergencyAddress(t *testing.T) {
	td := acceptance.NewTestData(t, "zoom_phone_emergency_address", "test")

	emergencyAddressTestResource := newEmergencyAddressResource(td)
	dataConfig := fmt.Sprintf(`
data "zoom_phone_emergency_address" "test" {
  # Load data after the resource is created using the resource
  id = %[1]s
}
`,
		fmt.Sprintf("%s.id", td.ResourceName),
	)
	expectedDataSourceName := "data.zoom_phone_emergency_address.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:             testAccPhoneEmergencyAddressResourceDestroy(td),
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: acceptance.ProviderConfig + emergencyAddressTestResource.requiredConfig() + dataConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedDataSourceName, "address_line1", emergencyAddressTestResource.AddressLine1),
					resource.TestCheckResourceAttr(expectedDataSourceName, "city", emergencyAddressTestResource.City),
					resource.TestCheckResourceAttr(expectedDataSourceName, "country", emergencyAddressTestResource.Country),
					resource.TestCheckResourceAttr(expectedDataSourceName, "state_code", emergencyAddressTestResource.StateCode),
					resource.TestCheckResourceAttr(expectedDataSourceName, "zip", emergencyAddressTestResource.Zip),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "status"),
					resource.TestCheckResourceAttrSet(expectedDataSourceName, "level"),
				),
			},
		},
	})
}
