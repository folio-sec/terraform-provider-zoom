---
# generated by https://github.com/hashicorp/terraform-plugin-docs with own template
page_title: "zoom_phone_call_handling_closed_hours Resource - zoom"
subcategory: "Phone"
description: |-
  Call handling settings allow you to control how your system routes calls during closed hours.
  For more information, read our Call Handling API guide https://developers.zoom.us/docs/zoom-phone/call-handling/ or Zoom support article Customizing call handling settings https://support.zoom.us/hc/en-us/articles/360059966372-Customizing-call-handling-settings.
  NOTE
  This resource is depends on zoom_phone_call_handling_business_hours. Please set business hours type = 2 (Custom hours).
  API Permissions
  The following API permissions are required in order to use this resource.
  This resource requires the phone:read:call_handling_setting:admin, phone:write:call_handling_setting:admin, phone:update:call_handling_setting:admin, phone:delete:call_handling_setting:admin.
---

# zoom_phone_call_handling_closed_hours (Resource)

Call handling settings allow you to control how your system routes calls during closed hours.
For more information, read our [Call Handling API guide](https://developers.zoom.us/docs/zoom-phone/call-handling/) or Zoom support article [Customizing call handling settings](https://support.zoom.us/hc/en-us/articles/360059966372-Customizing-call-handling-settings).

## NOTE
This resource is depends on `zoom_phone_call_handling_business_hours`. Please set business hours type = 2 (Custom hours).

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the `phone:read:call_handling_setting:admin`, `phone:write:call_handling_setting:admin`, `phone:update:call_handling_setting:admin`, `phone:delete:call_handling_setting:admin`.

## Example Usage

```terraform
# Define only one resource for each extension id.
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `call_handling` (Attributes) The call handling settings.
  - NOTE: some fields doesn't return from zoom api, so please ignore_changes for these fields. (see [below for nested schema](#nestedatt--call_handling))
- `extension_id` (String) Extension ID.

### Optional

- `call_forwarding` (Attributes) The call forwarding settings. (see [below for nested schema](#nestedatt--call_forwarding))

<a id="nestedatt--call_handling"></a>
### Nested Schema for `call_handling`

Optional:

- `allow_callers_check_voicemail` (Boolean) Whether to allow the callers to check voicemails over a phone. It's required only when the call_not_answer_action setting is set to 1 (Forward to a voicemail).
- `busy_forward_to_extension_id` (String) The forwarding extension ID that's required only when busy_on_another_call_action setting is set to:
  - 2 - Forward to the user.
  - 4 - Forward to the common area.
  - 6 - Forward to the auto receptionist.
  - 7 - Forward to a call queue.
  - 8 - Forward to a shared line group.
  - 9 - forward to an external contact.
- `busy_on_another_call_action` (Number) The action to take when the user is busy on another call:
  - 1 — Forward to a voicemail.
  - 2 — Forward to the user.
  - 4 — Forward to the common area.
  - 6 — Forward to the auto receptionist.
  - 7 — Forward to a call queue.
  - 8 — Forward to a shared line group.
  - 9 — Forward to an external contact.
  - 10 - Forward to a phone number.
  - 12 — Play a message, then disconnect.
  - 21 — Call waiting.
  - 22 — Play a busy signal.
- `busy_phone_number` (String) The extension's phone number or forward to an external number in [E.164](https://en.wikipedia.org/wiki/E.164) format format. It sets when `busy_on_another_call_action` action is set to `10` - Forward to an external number.
- `busy_phone_number_description` (String) This field forwards to an external number description (optional). It sets when `busy_on_another_call_action` action is set to `10` - Forward to an external number.
- `busy_play_callee_voicemail_greeting` (Boolean) Whether to play callee's voicemail greeting when the caller reaches the end of the forwarding sequence. It displays when busy_on_another_call_action action is set to
  - 2 - Forward to the user
  - 4 - Forward to the common area
  - 6 - Forward to the auto receptionist
  - 7 - Forward to a call queue
  - 8 - Forward to a shared line group
  - 9 - Forward to an external contact
  - 10 - Forward to an external number.
- `busy_require_press1_before_connecting` (Boolean) When one is busy on another call, the receiver needs to press 1 before connecting the call for it to be forwarded to an external contact or a number. This option ensures that forwarded calls won't reach the voicemail box for the external contact or a number.
- `call_not_answer_action` (Number) The action to take when a call is not answered:
  - 1 — Forward to a voicemail.
  - 2 — Forward to the user.
  - 4 — Forward to the common area.
  - 6 — Forward to the auto receptionist.
  - 7 — Forward to a call queue.
  - 8 — Forward to a shared line group.
  - 9 — Forward to an external contact.
  - 10 - Forward to a phone number.
  - 11 — Disconnect.
  - 12 — Play a message, then disconnect.
  - 13 - Forward to a message.
  - 14 - Forward to an interactive voice response (IVR).
- `connect_to_operator` (Boolean) Whether to allow callers to reach an operator. It's required only when the `call_not_answer_action` or `busy_on_another_call_action` is set to 1 (Forward to a voicemail).
- `forward_to_extension_id` (String) The forwarding extension ID that's required only when call_not_answer_action setting is set to:
  - 2 - Forward to the user.
  - 4 - Forward to the common area.
  - 6 - Forward to the auto receptionist.
  - 7 - Forward to a call queue.
  - 8 - Forward to a shared line group.
  - 9 - forward to an external contact.
- `greeting_prompt_id` (String) The greeting audio prompt ID.
  - Options: empty char - default and 0 - disable
  - This is only required for the Call Queue or Auto Receptionist call_handling sub-setting.
- `max_wait_time` (Number) The maximum wait time, in seconds.
  - for simultaneous ring mode or the ring duration for each device for sequential ring mode: 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60.
  - Specify how long a caller will wait in the queue. Once the wait time is exceeded, the caller will be rerouted based on the overflow option for Call Queue: 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 120, 180, 240, 300, 600, 900, 1200, 1500, 1800.
  - This is only required for the call_handling sub-setting.
- `operator_extension_id` (String) The extension ID of the operator to whom the call is being forwarded. It's required only when `call_not_answer_action` is set to `1` (Forward to a voicemail) and `connect_to_operator` is set to true.
- `overflow_play_callee_voicemail_greeting` (Boolean) Whether to play the callee's voicemail greeting when the caller reaches the end of the forwarding sequence. It displays when call_not_answer_action is set to:
  - 2 - Forward to the user
  - 4 - Forward to the common area
  - 6 - Forward to the auto receptionist
  - 7 - Forward to a call queue
  - 8 - Forward to a shared line group
  - 9 - Forward to an external contact
  - 10 - Forward to an external number.
- `phone_number` (String) The extension's phone number or forward to an external number in [E.164](https://en.wikipedia.org/wiki/E.164) format format. It's required when `call_not_answer_action` action is set to `10` - Forward to an external number.
- `phone_number_description` (String) (Optional) This field forwards to an external number description. Add this field when `call_not_answer_action` is set to `10` - Forward to an external number.
- `play_callee_voicemail_greeting` (Boolean) Whether to play callee's voicemail greeting when the caller reaches the end of forwarding sequence. It displays when `busy_on_another_call_action` action or `call_not_answer_action` is set to `1` - Forward to a voicemail.
- `ring_mode` (String) The call handling ring mode:
  - simultaneous
  - sequential. For user closed hours, ring_mode needs to be set with max_wait_time.
- `unanswered_require_press1_before_connecting` (Boolean) When a call is unanswered, press 1 before connecting the call to forward to an external contact or a number. This option ensures that forwarded calls won't reach the voicemail box for the external contact or a number.


<a id="nestedatt--call_forwarding"></a>
### Nested Schema for `call_forwarding`

Optional:

- `enable_zoom_desktop_apps` (Boolean) Whether to enable Zoom Desktop Apps call forwarding
- `enable_zoom_mobile_apps` (Boolean) Whether to enable Zoom Mobile Apps call forwarding
- `enable_zoom_phone_appliance_apps` (Boolean) Whether to enable Zoom Phone Appliance Apps call forwarding
- `require_press_1_before_connecting` (Boolean) When a call is forwarded to a personal phone number, whether the user must press "1" before the call connects. Enable this option to ensure missed calls do not reach to your personal voicemail. It's required for the `call_forwarding` sub-setting. Press 1 is always enabled and is required for callQueue type extension calls.
- `settings` (Attributes Set) The call forwarding settings. It's only required for the `call_forwarding` sub-setting. (see [below for nested schema](#nestedatt--call_forwarding--settings))

<a id="nestedatt--call_forwarding--settings"></a>
### Nested Schema for `call_forwarding.settings`

Optional:

- `description` (String) The external phone number's description.
- `enable` (Boolean) Whether to receive a call.
- `phone_number` (String) The external phone number in [E.164](https://en.wikipedia.org/wiki/E.164) format format.

Read-Only:

- `id` (String) The call forwarding's ID.

## Import

Import is supported using the following syntax:

```shell
# ${extension_id}
terraform import zoom_phone_call_handling_closed_hours.example t6wyhAZRQXXX_Rv3jj3XXX
```
