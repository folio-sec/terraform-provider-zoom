resource "zoom_phone_external_contact" "example" {
  id          = "4gAYcVCwTUGGDpYWdFpXxx"
  name        = "Example"
  description = "Description"
  email       = "example@example.com"
  phone_numbers = [
    "+0123456789",
  ]
  auto_call_recorded = true
}

resource "zoom_phone_external_contact" "other" {
  name = "Other"
  phone_numbers = [
    "+0123456789",
  ]
}
