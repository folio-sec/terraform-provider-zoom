resource "zoom_phone_call_queue" "example" {
  name             = "terraform-example"
  extension_number = "1234"
}

resource "zoom_phone_call_queue_members" "example" {
  call_queue_id = zoom_phone_call_queue.example.id
  common_areas = [
    {
      id = "xxxxx-Q6aYBcsv2wJaag"
    }
  ]
  users = [
    {
      id = "YYYgNJuS-XXcsv2wJnug"
    },
    {
      email = "mary@example.com"
    },
    {
      email = "john@example.com"
    },
  ]
}
