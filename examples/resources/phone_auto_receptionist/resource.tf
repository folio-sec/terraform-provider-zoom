locals {
  prefix = "example-terraform"
}

resource "zoom_phone_auto_receptionist" "example" {
  name = "${local.prefix}-example"
}
