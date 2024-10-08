---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hwmux_device_group Data Source - hwmux"
subcategory: ""
description: |-
  DeviceGroup data source
---

# hwmux_device_group (Data Source)

DeviceGroup data source



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (Number) Device Group identifier

### Read-Only

- `devices` (Attributes List) The devices that belong to the Device Group (see [below for nested schema](#nestedatt--devices))
- `enable_ahs` (Boolean) Enable the Automated Health Service
- `enable_ahs_actions` (Boolean) Allow the Automated Health Service to take DeviceGroups offline when they are unhealthy.
- `enable_ahs_cas` (Boolean) Enable the Automated Health Service to take Corrective Actions.
- `metadata` (String) The metadata of the Device Group.
- `name` (String) Device Group name. Must be unique.
- `source` (String) The source where the device group was created.

<a id="nestedatt--devices"></a>
### Nested Schema for `devices`

Read-Only:

- `id` (Number) Device ID.


