---
# generated by https://github.com/hashicorp/terraform-plugin-docs with own template
page_title: "zoom_phone_site Data Source - zoom"
subcategory: "Phone"
description: |-
  A Zoom Phone site in a Zoom account.
  API Permissions
  The following API permissions are required in order to use this resource.
  This resource requires the phone:read:site:admin.
---

# zoom_phone_site (Data Source)

A Zoom Phone site in a Zoom account.

## API Permissions

The following API permissions are required in order to use this resource.
This resource requires the `phone:read:site:admin`.

## Example Usage

```terraform
data "zoom_phone_site" "example" {
  id = "example-site-id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The site ID is the unique identifier of the site.

### Read-Only

- `caller_id_name` (String) When an outbound call uses a number as the caller ID, the caller ID name and the number display to the called party. The caller ID name can be up to 15 characters. The user can reset the caller ID name by setting it to empty string.
- `country` (Attributes) The country of the site. (see [below for nested schema](#nestedatt--country))
- `india_city` (String) The India site's city. This field only applies to India based accounts.
- `india_entity_name` (String) When select the Indian sip zone, then need to set the entity name. This field only applies to India based accounts.
- `india_sdca_npa` (String) The India site's Short Distance Calling Area (sdca) Numbering Plan Area (npa). This field is linked to the 'state_code' field. This field only applies to India based accounts.
- `india_state_code` (String) The India site's state code. This field only applies to India based accounts.
- `level` (String) The level of the site.
- `main_auto_receptionist` (Attributes) The [main auto receptionist](https://support.zoom.us/hc/en-us/articles/360021121312#h_bc7ff1d5-0e6c-40cd-b889-62010cb98c57) for each site. (see [below for nested schema](#nestedatt--main_auto_receptionist))
- `name` (String) The name of the site.
- `short_extension` (Attributes) The short extension of the phone site. (see [below for nested schema](#nestedatt--short_extension))
- `sip_zone` (Attributes) When you select a SIP zone nearest to your site, it might help reduce latency and improve call quality. (see [below for nested schema](#nestedatt--sip_zone))
- `site_code` (Number) The [site code](https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0069806).

<a id="nestedatt--country"></a>
### Nested Schema for `country`

Read-Only:

- `code` (String) The two lettered country [code](https://developers.zoom.us/docs/api/references/abbreviations/).
- `name` (String) The name of the country.


<a id="nestedatt--main_auto_receptionist"></a>
### Nested Schema for `main_auto_receptionist`

Read-Only:

- `id` (String) The auto receptionist ID.
- `name` (String) Display name of the [auto-receptionist](https://support.zoom.com/hc/en/article?id=zm_kb&sysparm_article=KB0061421) as main auto-receptionist for the site.


<a id="nestedatt--short_extension"></a>
### Nested Schema for `short_extension`

Read-Only:

- `length` (Number) This setting specifies the length of short extension numbers for the site. The value must be between 1 and 6., Default is `3`.


<a id="nestedatt--sip_zone"></a>
### Nested Schema for `sip_zone`

Read-Only:

- `id` (String) The SIP zone ID.
- `name` (String) The SIP zone name.
