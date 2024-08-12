data "zoom_phone_auto_receptionist" "primary" {
  id = "I-fiz4v1TFC4FWfzXA0XXX"
}

resource "zoom_phone_auto_receptionist" "example" {
  name = "terraform-example"
}

resource "zoom_phone_auto_receptionist_ivr" "example" {
  auto_receptionist_id = zoom_phone_auto_receptionist.example.id
  hours_type           = "business_hours"
  caller_enters_no_action = {
    audio_prompt_repeat = 1
    action              = 6 # Forward to the auto receptionist
    forward_to = {
      extension_id = data.zoom_phone_auto_receptionist.primary.extension_id
    }
  }
  key_actions = {
    "0" = {
      action = 100 # Leave voicemail to the current extension
      voicemail_greeting = {
        id = "" # empty string as using default audio
      }
    },
    "1" = {
      action = 6 # Forward to the auto receptionist
      target = {
        extension_id = zoom_phone_auto_receptionist.example.extension_id
      }
    },
    "*" = {
      action = "-1" # Disabled
    },
    "#" = {
      action = "21" # Repeat menu greeting
    }
  }
}
