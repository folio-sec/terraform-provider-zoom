# Define only one resource for each extended id.
resource "zoom_phone_call_handling_closed_hours" "example" {
  extension_id = "wGJDBcnJQC6tV86BbtlXXX"

  call_handling = {
    call_not_answer_action        = 1
    busy_on_another_call_action   = 21
    connect_to_operator           = false
    allow_callers_check_voicemail = false
    voicemail_greeting_id         = "" # default
    ring_mode                     = "simultaneous"
    max_wait_time                 = 30
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

  # closed hours can handle after setting business_hours.custom_hours.type = 2
  depends_on = [
    zoom_phone_call_handling_business_hours.example
  ]

  lifecycle {
    ignore_changes = [
      # zoom api doesn't return some fields, so please ignore them
      call_handling.receive_call,
    ]
  }
}
