resource "zoom_phone_site" "example" {
  name = "example-site"

  main_auto_receptionist = {
    name = "example-auto-receptionist"
  }

  default_emergency_address = {
    address_line1 = "123 Main St"
    address_line2 = "Suite 100"
    city          = "San Jose"
    country       = "US"
    state_code    = "CA"
    zip           = "95131"
  }
}
