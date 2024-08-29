data "zoom_phone_users" "example" {
  query = {
    status = "activate"
  }
}

output "users" {
  value = data.zoom_phone_users.example.users
}
