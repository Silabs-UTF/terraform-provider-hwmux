resource "hwmux_device_group" "new_testbed" {
  name              = "new_testbed"
  metadata          = jsonencode(yamldecode(file("example.yaml")))
  devices           = [1, 2]
  permission_groups = ["Example group name"]
}
