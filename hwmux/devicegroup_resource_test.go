package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDeviceGroupResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: providerConfig + `
resource "hwmux_deviceGroup" "test" {
	name     = "test_dg"
	devices = [1, 2]
	permission_groups = ["All users"]
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "name", "test_dg"),
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "devices.#", "2"),
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "devices.0", "1"),
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "permission_groups.0", "All users"),
                    // Verify the deviceGroup item has Computed attributes filled.
                    resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "metadata", "{}"),
                    // Verify dynamic values have any value set in the state.
                    resource.TestCheckResourceAttrSet("hwmux_deviceGroup.test", "id"),
                    resource.TestCheckResourceAttrSet("hwmux_deviceGroup.test", "last_updated"),
                ),
            },
            // ImportState testing
            {
                ResourceName:      "hwmux_deviceGroup.test",
                ImportState:       true,
                ImportStateVerify: true,
                // The last_updated attribute does not exist in the HashiCups
                // API, therefore there is no value for it during import.
                ImportStateVerifyIgnore: []string{"last_updated"},
            },
            // Update and Read testing
            {
                Config: providerConfig + `
resource "hwmux_deviceGroup" "test" {
	name     = "test_dg"
	devices = [1]
	permission_groups = ["Staff users"]
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "name", "test_dg"),
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "devices.#", "1"),
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "devices.0", "1"),
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "permission_groups.0", "Staff users"),
                    // Verify the deviceGroup item has Computed attributes filled.
                    resource.TestCheckResourceAttr("hwmux_deviceGroup.test", "metadata", "{}"),
                ),
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}
