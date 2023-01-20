resource "hwmux_device_group" "new_testbed" {
  name              = "name_no_spaces"
  metadata          = jsonencode(yamldecode(file("example.yaml")))
  devices           = [1, 2]
  permission_groups = ["Example group name"]
}
