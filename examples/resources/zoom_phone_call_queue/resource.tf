resource "zoom_phone_call_queue" "example" {
  name             = "terraform-example"
  extension_number = "123"
}

resource "zoom_phone_call_queue" "inactive" {
  name             = "terraform-example-inactive"
  extension_number = "124"
  description      = "Example description"
  cost_center      = "XXX"
  department       = "YYY"
  status           = "inactive"
}
