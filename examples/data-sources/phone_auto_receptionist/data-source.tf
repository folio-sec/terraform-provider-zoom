data "zoom_phone_auto_receptionist" "example" {
  auto_receptionist_id = "example"
}

output "example" {
  value = data.zoom_phone_auto_receptionist.example
}
