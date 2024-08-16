resource "zoom_phone_call_queue" "example" {
  name             = "terraform-example"
  extension_number = "1234"
}

resource "zoom_phone_call_queue_members" "example" {
  call_queue_id = zoom_phone_call_queue.example.id
  common_areas  = []
  users = [
    {
      id    = "6KpvKpy-RFCYmhj-XXXFqA"
      email = "john@folio-sec.com"
    },
  ]
}
