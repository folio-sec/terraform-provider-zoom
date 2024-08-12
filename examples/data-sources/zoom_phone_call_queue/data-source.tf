data "zoom_phone_call_queue" "example" {
  id = "wGJDBcnJQC6tV86Bbtlq1Q"
}

output "example" {
  value = data.zoom_phone_call_queue.example
}
