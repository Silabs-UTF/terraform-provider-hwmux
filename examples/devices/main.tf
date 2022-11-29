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

output "sn0_device" {
  value = data.hwmux_device.sn0
}
