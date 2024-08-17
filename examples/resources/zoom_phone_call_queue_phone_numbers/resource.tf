resource "zoom_phone_call_queue" "example" {
  name             = "terraform-example"
  extension_number = "1234"
}

resource "zoom_phone_call_queue_phone_numbers" "example" {
  call_queue_id = zoom_phone_call_queue.example.id
  phone_numbers = [
    {
      id = "gFARuKuQQ2qmR4ldyQrViQ",
      # number = "+12058945456",
    },
  ]
}
