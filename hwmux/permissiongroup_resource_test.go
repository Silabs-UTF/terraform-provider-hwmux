package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPermissionGroupResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: providerConfig + `
resource "hwmux_permission_group" "test" {
	name     = "test_permission_group_tf"
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("hwmux_permission_group.test", "name", "test_permission_group_tf"),
                    resource.TestCheckResourceAttrSet("hwmux_permission_group.test", "id"),
                    resource.TestCheckResourceAttrSet("hwmux_permission_group.test", "last_updated"),
                ),
            },
            // Update and Read testing
            {
                Config: providerConfig + `
resource "hwmux_permission_group" "test" {
    name     = "test_permission_group_tf-2"
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_permission_group.test", "name", "test_permission_group_tf-2"),
                ),
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}
