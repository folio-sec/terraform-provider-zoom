# Define multiple resources for each extended id.
resource "zoom_phone_call_handling_holiday_hours" "example" {
  extension_id = "wGJDBcnJQC6tV86BbtlXXX"

  holiday = {
    name = "temporary"
    from = "2024-08-24T09:00:00Z"
    to   = "2024-08-26T17:00:00+09:00"
  }

  call_handling = {
    call_not_answer_action        = 1
    connect_to_operator           = false
    allow_callers_check_voicemail = false
  }

  # some extension can use call forwarding such as user
  call_forwarding = {
    require_press_1_before_connecting = true
    enable_zoom_mobile_apps           = true # Zoom Mobile Apps
    enable_zoom_desktop_apps          = true # Zoom Desktop Apps
    enable_zoom_phone_appliance_apps  = true # Zoom Phone Appliance Apps
    settings = [
      {
        "description" : "external person",
        "enable" : true,
        "phone_number" : "+1234567890"
      },
    ]
  }

  lifecycle {
    ignore_changes = [
      # zoom api doesn't return some fields, so please ignore them
      call_handling.receive_call,
    ]
  }
}
