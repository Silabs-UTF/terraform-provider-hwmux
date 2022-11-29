terraform {
  required_providers {
    hwmux = {
      source = "silabs.com/iot-infra-sw/hwmux"
    }
  }
}

provider "hwmux" {
  host  = "http://localhost"
  token = "cd48869f98f4eca6eb3bda542f5ac808e4beabc4"
}

data "hwmux_device" "sn0" {
  id = 1
}

data "hwmux_deviceGroup" "testbed1" {
  id = 1
}

resource "hwmux_device" "new_device" {
  sn_or_name = "new device"
  uri        = "99.9.9.1"
  part       = "Part_no_0"
  room       = "Room_0"
  metadata   = jsonencode(yamldecode(file("example.yaml")))
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
