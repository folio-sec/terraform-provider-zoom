data "zoom_phone_shared_line_group" "example" {
  id = "4gAYcVCwTUGGDpYWdFpXxx"
}

output "example" {
  value = data.zoom_phone_shared_line_group.example
}
