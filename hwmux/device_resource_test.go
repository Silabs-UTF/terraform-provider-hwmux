package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDeviceResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: providerConfig + `
resource "hwmux_device" "test" {
	sn_or_name = "test_device"
    uri = "77.7.7.7"
    part = "Part_no_0"
    room = "Room_0"
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("hwmux_device.test", "sn_or_name", "test_device"),
					resource.TestCheckResourceAttr("hwmux_device.test", "uri", "77.7.7.7"),
					resource.TestCheckResourceAttr("hwmux_device.test", "part", "Part_no_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "room", "Room_0"),
                    // Verify the device item has Computed attributes filled.
                    resource.TestCheckResourceAttr("hwmux_device.test", "is_wstk", "false"),
                    resource.TestCheckResourceAttr("hwmux_device.test", "online", "true"),
                    resource.TestCheckResourceAttr("hwmux_device.test", "metadata", "{}"),
					resource.TestCheckResourceAttr("hwmux_device.test", "location_metadata", "{}"),
                    // Verify dynamic values have any value set in the state.
                    resource.TestCheckResourceAttrSet("hwmux_device.test", "id"),
                    resource.TestCheckResourceAttrSet("hwmux_device.test", "last_updated"),
                ),
            },
            // ImportState testing
            {
                ResourceName:      "hwmux_device.test",
                ImportState:       true,
                ImportStateVerify: true,
                // The last_updated attribute does not exist in the HashiCups
                // API, therefore there is no value for it during import.
                ImportStateVerifyIgnore: []string{"last_updated"},
            },
            // Update and Read testing
            {
                Config: providerConfig + `
resource "hwmux_device" "test" {
	sn_or_name = "test_device"
	uri = "88.8.8.8"
	part = "Part_no_0"
	room = "Room_0"
	online = false
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Verify first device item updated
                    resource.TestCheckResourceAttr("hwmux_device.test", "sn_or_name", "test_device"),
					resource.TestCheckResourceAttr("hwmux_device.test", "uri", "88.8.8.8"),
					resource.TestCheckResourceAttr("hwmux_device.test", "part", "Part_no_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "room", "Room_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "online", "false"),
                    // Verify first coffee item has Computed attributes updated.
				   resource.TestCheckResourceAttr("hwmux_device.test", "is_wstk", "false"),
				   resource.TestCheckResourceAttr("hwmux_device.test", "metadata", "{}"),
				   resource.TestCheckResourceAttr("hwmux_device.test", "location_metadata", "{}"),
                ),
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}
