resource "zoom_phone_shared_line_group" "example" {
  display_name     = "terraform-example"
  extension_number = "1234"
}

resource "zoom_phone_shared_line_group" "inactive" {
  display_name     = "terraform-example-inactive"
  extension_number = "1235"
  status           = "inactive"
}
