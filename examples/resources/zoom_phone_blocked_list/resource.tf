resource "zoom_phone_blocked_list" "example" {
  block_type   = "inbound"
  match_type   = "phoneNumber"
  phone_number = "+1234567890"
  status       = "active"
}
