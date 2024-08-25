# Define only one resource for each extended id.
resource "zoom_phone_call_handling_business_hours" "example" {
  extension_id = "wGJDBcnJQC6tV86BbtlXXX"

  custom_hours = {
    type                   = 2 # Custom hours
    allow_members_to_reset = true
    settings = [
      {
        weekday = 1 # Sunday
        type    = 0 # Disabled
      },
      {
        weekday = 2 # Monday
        type    = 2 # Customized hours
        from    = "09:00"
        to      = "22:00"
      },
      {
        weekday = 3 # Saturday
        type    = 2 # Customized hours
        from    = "09:00"
        to      = "23:59"
      },
      {
        weekday = 4 # Saturday
        type    = 1 # 24hours
      },
      {
        weekday = 5 # Saturday
        type    = 1 # 24hours
      },
      {
        weekday = 3 # Saturday
        type    = 2 # Customized hours
        from    = "00:00"
        to      = "22:00"
      },
      {
        weekday = 7 # Saturday
        type    = 0 # Disabled
      },
    ]
  }

  call_handling = {
    call_not_answer_action        = 7
    forward_to_extension_id       = "XXX"
    busy_on_another_call_action   = 21
    busy_forward_to_extension_id  = "XXX"
    allow_callers_check_voicemail = true
    allow_members_to_reset        = false
    audio_while_connecting_id     = "XXX"
    call_distribution = {
      handle_multiple_calls            = true
      ring_duration                    = 30
      ring_mode                        = "simultaneous"
      skip_offline_device_phone_number = true
    }
    busy_require_press_1_before_connecting        = true
    un_answered_require_press_1_before_connecting = true
    overflow_play_callee_voicemail_greeting       = true
    play_callee_voicemail_greeting                = true
    busy_play_callee_voicemail_greeting           = true
    phone_number                                  = "+1234567890"
    phone_number_description                      = "XXX"
    busy_phone_number                             = "+1234567890"
    busy_phone_number_description                 = ""
    connect_to_operator                           = true
    greeting_prompt_id                            = "0" # default
    max_call_in_queue                             = 20
    max_wait_time                                 = 30
    music_on_hold_id                              = "0" # default
    operator_extension_id                         = "XXX"
    receive_call                                  = true
    ring_mode                                     = "simultaneous"
    voicemail_greeting_id                         = ""
    wrap_up_time                                  = 60
  }

  # some extension can use call forwarding such as user
  call_forwarding = {
    require_press_1_before_connecting = true
    enable_zoom_mobile_apps           = true # Zoom Mobile Apps
    enable_zoom_desktop_apps          = true # Zoom Desktop Apps
    enable_zoom_phone_appliance_apps  = true # Zoom Phone Appliance Apps
    settings = [
      {
        "description"  = "external person",
        "enable"       = false,
        "phone_number" = "+1234567890"
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
