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
            // Update and Read testing
            {
                Config: providerConfig + `
resource "hwmux_user" "test" {
    username     = "team1_jenkins-2"
	permission_groups = ["Staff users"]
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_user.test", "username", "team1_jenkins-2"),
                ),
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}
