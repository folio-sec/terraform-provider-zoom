---
# generated by https://github.com/hashicorp/terraform-plugin-docs with own template
page_title: "zoom_phone_call_queue_phone_numbers Resource - zoom"
subcategory: "Phone"
description: |-
  After buying phone number(s) https://support.zoom.us/hc/en-us/articles/360020808292#h_007ec8c2-0914-4265-8351-96ab23efa3ad, you can assign it, allowing callers to directly dial a number to reach a call queue https://support.zoom.us/hc/en-us/articles/360021524831-Managing-Call-Queues.
  API Permissions
  The following API permissions are required in order to use this resource.
  This resource requires the phone:read:call_queue:admin, phone:read:list_call_queues:admin, phone:read:list_numbers:admin, phone:write:call_queue_number:admin, phone:delete:call_queue_number:admin.
---

# zoom_phone_call_queue_phone_numbers (Resource)

After [buying phone number(s)](https://support.zoom.us/hc/en-us/articles/360020808292#h_007ec8c2-0914-4265-8351-96ab23efa3ad), you can assign it, allowing callers to directly dial a number to reach a [call queue](https://support.zoom.us/hc/en-us/articles/360021524831-Managing-Call-Queues).

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the `phone:read:call_queue:admin`, `phone:read:list_call_queues:admin`, `phone:read:list_numbers:admin`, `phone:write:call_queue_number:admin`, `phone:delete:call_queue_number:admin`.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `call_queue_id` (String) Unique identifier of the Call Queue.
- `phone_numbers` (Attributes Set) (see [below for nested schema](#nestedatt--phone_numbers))

<a id="nestedatt--phone_numbers"></a>
### Nested Schema for `phone_numbers`

Optional:

- `id` (String) Unique identifier of the number. Provide either the `id` or the `number` field.
- `number` (String) Phone number e.g. `+12058945456`. Provide either the `id` or the `number` field.
- `source` (String) Source
  - Allowed: internal┃external

## Import

Import is supported using the following syntax:

```shell
# ${call_queue_id}
terraform import zoom_phone_call_queue_phone_numbers.example wGJDBcnJQC6tV86BbtlXXX
```
