resource "zoom_phone_shared_line_group" "example" {
  display_name     = "terraform-example"
  extension_number = "1234"
}

resource "zoom_phone_shared_line_group_phone_numbers" "example" {
  shared_line_group_id = zoom_phone_shared_line_group.example.id
  primary_number       = "+1234567890"
  phone_numbers = [
    {
      number = "+1234567890",
    },
    {
      id = "4gAYcVCwTUGGDpYWdFpXxx",
    },
  ]
}
