---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hwmux_device_group Resource - hwmux"
subcategory: ""
description: |-
  Device Group resource.
---

# hwmux_device_group (Resource)

Device Group resource.

## Example Usage

```terraform
resource "hwmux_device_group" "new_testbed" {
  name               = "name_no_spaces"
  metadata           = jsonencode(yamldecode(file("example.yaml")))
  devices            = [1, 2]
  permission_groups  = ["Example group name"]
  enable_ahs         = true
  enable_ahs_actions = true
  enable_ahs_cas     = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `devices` (Set of Number) The devices that belong to the Device Group.
- `name` (String) Device Group name.
- `permission_groups` (Set of String) Which permission groups can access the resource.

### Optional

- `enable_ahs` (Boolean) Enable the Automated Health Service
- `enable_ahs_actions` (Boolean) Allow the Automated Health Service to take DeviceGroups offline when they are unhealthy.
- `enable_ahs_cas` (Boolean) Allow the Automated Health Service to take corrective actions.
- `metadata` (String) The metadata of the Device Group.

### Read-Only

- `id` (String) Device Group identifier.
- `last_updated` (String) Timestamp of the last Terraform update of the resource.
- `source` (String) The source where the device group was created.

## Import

Import is supported using the following syntax:

```shell
# A deviceGroup can be imported by specifying its ID
terraform import hwmux_device_group.example 123
```
