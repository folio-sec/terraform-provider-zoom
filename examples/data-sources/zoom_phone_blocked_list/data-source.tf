data "zoom_phone_blocked_list" "example" {
  id = "lSq8jyDORe6tmbaUkOVhXx"
}

output "example" {
  value = data.zoom_phone_blocked_list.example
}
