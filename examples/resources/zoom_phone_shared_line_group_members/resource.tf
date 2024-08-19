resource "zoom_phone_shared_line_group" "example" {
  display_name     = "terraform-example"
  extension_number = "1234"
}

resource "zoom_phone_shared_line_group_members" "example" {
  shared_line_group_id = zoom_phone_shared_line_group.example.id
  common_areas = [
    {
      id = "xxxxx-Q6aYBcsv2wJaag"
    }
  ]
  users = [
    {
      id = "4gAYcVCwTUGGDpYWdFpXxx"
    },
    {
      email = "mary@example.com"
    },
    {
      email = "john@example.com"
    },
  ]
}
