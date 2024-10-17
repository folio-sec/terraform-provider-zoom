resource "zoom_phone_call_queue" "example" {
  name             = "terraform-example"
  extension_number = "1234"
}

resource "zoom_phone_call_queue_policy_voice_mail" "example" {
  call_queue_id = zoom_phone_call_queue.example.id

  access_members = [
    {
      access_user_id = "LLgNJuS-Q6aYBcsv2wJnug", # Zoom User Id (not phone user id)
      allow_download = true
      allow_delete   = false
      allow_sharing  = true
    },
  ]
}
