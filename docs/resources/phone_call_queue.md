---
# generated by https://github.com/hashicorp/terraform-plugin-docs with own template
page_title: "zoom_phone_call_queue Resource - zoom"
subcategory: "Phone"
description: |-
  Call queues allow you to route incoming calls to a group of users. For instance, you can use call queues to route calls to various departments in your organization such as sales, engineering, billing, customer service etc.
  API Permissions
  The following API permissions are required in order to use this resource.
  This resource requires the phone:read:call_queue:admin, phone:write:call_queue:admin, phone:update:call_queue:admin, phone:delete:call_queue:admin.
---

# zoom_phone_call_queue (Resource)

Call queues allow you to route incoming calls to a group of users. For instance, you can use call queues to route calls to various departments in your organization such as sales, engineering, billing, customer service etc.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the `phone:read:call_queue:admin`, `phone:write:call_queue:admin`, `phone:update:call_queue:admin`, `phone:delete:call_queue:admin`.

## Example Usage

```terraform
resource "zoom_phone_call_queue" "example" {
  name             = "terraform-example"
  extension_number = "123"
}

resource "zoom_phone_call_queue" "inactive" {
  name             = "terraform-example-inactive"
  extension_number = "124"
  description      = "Example description"
  cost_center      = "XXX"
  department       = "YYY"
  status           = "inactive"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `extension_number` (Number) Extension number of the call queue.
- `name` (String) Name of the call queue.

### Optional

- `cost_center` (String) Cost center name.
- `department` (String) Department name.
- `description` (String) Description for the Call Queue.
- `site_id` (String) The unique identifier of the site. It's required only if [multiple sites](https://support.zoom.us/hc/en-us/articles/360020809672-Managing-Multiple-Sites) have been enabled. This can be retrieved from the [List Phone Sites](https://marketplace.zoom.us/docs/api-reference/phone/methods#operation/listPhoneSites) API.
- `status` (String) Status of the Call Queue.
  - Allowed: active┃inactive

### Read-Only

- `extension_id` (String) Extension ID.
- `id` (String) Unique identifier of the Call Queue.

## Import

Import is supported using the following syntax:

```shell
# ${call_queue_id}
terraform import zoom_phone_call_queue.example wGJDBcnJQC6tV86BbtlXXX
```
