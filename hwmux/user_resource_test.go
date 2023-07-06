package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "hwmux_user" "test" {
	username     = "team1_jenkins"
    password     = "a_password"
	permission_groups = ["All users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_user.test", "username", "team1_jenkins"),
					resource.TestCheckResourceAttr("hwmux_user.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_user.test", "permission_groups.0", "All users"),
					resource.TestCheckResourceAttrSet("hwmux_user.test", "id"),
					resource.TestCheckResourceAttrSet("hwmux_user.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hwmux_user.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated", "password"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "hwmux_user" "test" {
    username     = "team1_jenkins-2"
    password     = "a_password-2"
	permission_groups = ["Staff users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_user.test", "username", "team1_jenkins-2"),
					resource.TestCheckResourceAttr("hwmux_user.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_user.test", "permission_groups.0", "Staff users"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
