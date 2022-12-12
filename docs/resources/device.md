---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hwmux_device Resource - hwmux"
subcategory: ""
description: |-
  Manages a device in hwmux.
---

# hwmux_device (Resource)

Manages a device in hwmux.

## Example Usage

```terraform
# Manage example device.
resource "hwmux_device" "example" {
  sn_or_name = "new device"
  uri        = "99.9.9.1"
  part       = "Part_no_0"
  room       = "Room_0"
  is_wstk    = false
  online     = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `part` (String) The part number of the device.
- `room` (String) The name of the room the device is in. Must exist in hwmux.

### Optional

- `is_wstk` (Boolean) Whether the device is a WSTK.
- `location_metadata` (String) The location metadata of the device.
- `metadata` (String) The metadata of the device.
- `online` (Boolean) Whether the device is online.
- `sn_or_name` (String) The name of the device. Must be unique.
- `uri` (String) The URI or IP address of the device.

### Read-Only

- `id` (String) The ID of the device.
- `last_updated` (String) Timestamp of the last Terraform update of the device.

## Import

Import is supported using the following syntax:

```shell
# A device can be imported by specifying its ID
terraform import hwmux_device.example 123
```