package emergencyaddress_test

import "fmt"

type emergencyAddressResource struct {
	ResourceLabel         string
	TerraformResourceType string
	AddressLine1          string
	City                  string
	Country               string
	StateCode             string
	Zip                   string
	SiteID                string
}

func (r emergencyAddressResource) copyWithAddressLine1(value string) emergencyAddressResource {
	return emergencyAddressResource{
		ResourceLabel:         r.ResourceLabel,
		TerraformResourceType: r.TerraformResourceType,
		AddressLine1:          value,
		City:                  r.City,
		Country:               r.Country,
		StateCode:             r.StateCode,
		Zip:                   r.Zip,
		SiteID:                r.SiteID,
	}
}

func (r emergencyAddressResource) requiredConfig() string {
	return fmt.Sprintf(`
resource %[1]q %[2]q {
  address_line1 = %[3]q
  city = %[4]q
  country = %[5]q
  state_code = %[6]q
  zip = %[7]q
	site_id = %[8]q
}
`,
		r.TerraformResourceType,
		r.ResourceLabel,
		r.AddressLine1,
		r.City,
		r.Country,
		r.StateCode,
		r.Zip,
		r.SiteID,
	)
}
