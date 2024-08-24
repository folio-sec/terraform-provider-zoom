resource "zoom_phone_user" "example" {
  user_id          = "AAggbuS-Q6aXXcsv2wnug"
  extension_number = 101
}

resource "zoom_phone_user_calling_plans" "example" {
  user_id = zoom_phone_user.example.user_id
  calling_plans = [
    {
      type = 207
    },
  ]
}
