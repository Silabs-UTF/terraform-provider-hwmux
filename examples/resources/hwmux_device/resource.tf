# Manage example device.
resource "hwmux_device" "example" {
  sn_or_name        = "new device"
  uri               = "99.9.9.1"
  part              = "Part_no_0"
  room              = "Room_0"
  is_wstk           = false
  online            = true
  permission_groups = ["Example group name"]
}
