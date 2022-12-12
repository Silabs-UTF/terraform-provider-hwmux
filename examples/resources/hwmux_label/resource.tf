resource "hwmux_label" "new_label" {
  name              = "new_label"
  metadata          = jsonencode(yamldecode(file("example.yaml")))
  device_groups     = [1, 2]
  permission_groups = ["Example group name"]
}
