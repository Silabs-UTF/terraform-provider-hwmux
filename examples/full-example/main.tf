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

data "hwmux_device" "sn0" {
  id = 1
}

data "hwmux_deviceGroup" "testbed1" {
  id = 1
}

data "hwmux_label" "label1" {
  id = 1
}

resource "hwmux_device" "new_device" {
  sn_or_name = "new device"
  uri        = "99.9.9.1"
  part       = "Part_no_0"
  room       = "Room_0"
  metadata   = jsonencode(yamldecode(file("example.yaml")))
}

resource "hwmux_deviceGroup" "new_testbed" {
  name              = "new_testbed"
  metadata          = jsonencode(yamldecode(file("example.yaml")))
  devices           = [hwmux_device.new_device.id, data.hwmux_device.sn0.id]
  permission_groups = ["All users"]
}

resource "hwmux_label" "new_label" {
  name              = "new_label"
  metadata          = jsonencode(yamldecode(file("example.yaml")))
  device_groups     = [hwmux_deviceGroup.new_testbed.id, data.hwmux_deviceGroup.testbed1.id]
  permission_groups = ["All users"]
}

output "sn0_device" {
  value = data.hwmux_device.sn0
}

output "new_device" {
  value = hwmux_device.new_device
}

output "testbed1" {
  value = data.hwmux_deviceGroup.testbed1
}

output "new_testbed" {
  value = hwmux_deviceGroup.new_testbed
}

output "label1" {
  value = data.hwmux_label.label1
}

output "new_label" {
  value = hwmux_label.new_label
}
