terraform {
  required_providers {
    hwmux = {
      source = "silabs.com/iot-infra-sw/hwmux"
    }
  }
}

provider "hwmux" {
  host  = "http://localhost"
  token = "6cbeb43325c187390ec505d3fff1d8488bfb806a"
}

data "hwmux_device" "device_ds" {
  id = 1
}

data "hwmux_device_group" "device_group_ds" {
  id = 1
}

data "hwmux_label" "label_ds" {
  id = 1
}

data "hwmux_permission_group" "permission_group_ds" {
  name = "Staff users"
}

data "hwmux_part" "part_ds" {
  part_no = "Part_no_0"
}

data "hwmux_room" "room_ds" {
  name = "Room_0"
}

resource "hwmux_device" "new_device" {
  sn_or_name        = "new device"
  uri               = "99.9.9.1"
  part              = "Part_no_0"
  room              = "Room_0"
  metadata          = jsonencode(yamldecode(file("example.yaml")))
  permission_groups = ["All users"]
}

resource "hwmux_device_group" "new_testbed" {
  name              = "new_testbed"
  metadata          = jsonencode(yamldecode(file("example.yaml")))
  devices           = [hwmux_device.new_device.id, data.hwmux_device.device_ds.id]
  permission_groups = ["All users"]
}

resource "hwmux_label" "new_label" {
  name              = "new_label"
  metadata          = jsonencode(yamldecode(file("example.yaml")))
  device_groups     = [hwmux_device_group.new_testbed.id, data.hwmux_device_group.device_group_ds.id]
  permission_groups = ["All users"]
}

resource "hwmux_permission_group" "new_permission_group" {
  name = "IC team"
}

output "device_data" {
  value = data.hwmux_device.device_ds
}

output "new_device" {
  value = hwmux_device.new_device
}

output "device_group_data" {
  value = data.hwmux_device_group.device_group_ds
}

output "new_testbed" {
  value = hwmux_device_group.new_testbed
}

output "label_data" {
  value = data.hwmux_label.label_ds
}

output "new_label" {
  value = hwmux_label.new_label
}

output "permission_group_data" {
  value = data.hwmux_permission_group.permission_group_ds
}

output "part_data" {
  value = data.hwmux_part.part_ds
}

output "room_data" {
  value = data.hwmux_room.room_ds
}
