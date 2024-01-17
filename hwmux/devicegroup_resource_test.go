package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var deviceGroupResourceTfName string = "hwmux_device_group.test"

func TestAccDeviceGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "hwmux_device_group" "test" {
	name     = "test_dg"
	devices = [1, 2]
	permission_groups = ["All users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "name", "test_dg"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "devices.#", "2"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "devices.0", "1"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "permission_groups.#", "1"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "permission_groups.0", "All users"),
					// Verify the device_group item has Computed attributes filled.
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "metadata", "{}"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet(deviceGroupResourceTfName, "id"),
					resource.TestCheckResourceAttrSet(deviceGroupResourceTfName, "enable_ahs"),
					resource.TestCheckResourceAttrSet(deviceGroupResourceTfName, "enable_ahs_actions"),
					resource.TestCheckResourceAttrSet(deviceGroupResourceTfName, "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      deviceGroupResourceTfName,
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "hwmux_device_group" "test" {
	name     = "test_dg"
	devices = [1]
	enable_ahs = true
	enable_ahs_actions = true
	enable_ahs_cas = true
	permission_groups = ["Staff users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "name", "test_dg"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "devices.#", "1"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "devices.0", "1"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "permission_groups.#", "1"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "permission_groups.0", "Staff users"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "enable_ahs", "true"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "enable_ahs_actions", "true"),
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "enable_ahs_cas", "true"),
					// Verify the device_group item has Computed attributes filled.
					resource.TestCheckResourceAttr(deviceGroupResourceTfName, "metadata", "{}"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
