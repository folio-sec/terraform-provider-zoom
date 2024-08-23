data "zoom_user_users" "example" {
  status = "active"
}

output "users" {
  value = data.zoom_user_users.example.users
  // Thee host_key attribute is sensitive, so it is marked as such to root output
  sensitive = true
}
