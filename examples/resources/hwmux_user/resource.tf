resource "hwmux_user" "new_user" {
  username          = "new user"
  password          = "a password"
  permission_groups = ["A permission group"]
}