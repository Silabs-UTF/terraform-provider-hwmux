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
    permission_groups = ["Staff users"]
	socketed_chip = "BRD1019A001"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_device.test", "sn_or_name", "test_device"),
					resource.TestCheckResourceAttr("hwmux_device.test", "uri", "77.7.7.7"),
					resource.TestCheckResourceAttr("hwmux_device.test", "part", "Part_no_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "room", "Room_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_device.test", "permission_groups.0", "Staff users"),
					// Verify the device item has Computed attributes filled.
					resource.TestCheckResourceAttr("hwmux_device.test", "is_wstk", "false"),
					resource.TestCheckResourceAttr("hwmux_device.test", "online", "true"),
					resource.TestCheckResourceAttr("hwmux_device.test", "metadata", "{}"),
					resource.TestCheckResourceAttr("hwmux_device.test", "location_metadata", "{}"),
					resource.TestCheckResourceAttr("hwmux_device.test", "source", "TERRAFORM"),
					resource.TestCheckResourceAttr("hwmux_device.test", "socketed_chip", "BRD1019A001"),
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
	socketed_chip = "BRD1019A002"
    permission_groups = ["All users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify first device item updated
					resource.TestCheckResourceAttr("hwmux_device.test", "sn_or_name", "test_device"),
					resource.TestCheckResourceAttr("hwmux_device.test", "uri", "88.8.8.8"),
					resource.TestCheckResourceAttr("hwmux_device.test", "part", "Part_no_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "room", "Room_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "online", "false"),
					resource.TestCheckResourceAttr("hwmux_device.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_device.test", "permission_groups.0", "All users"),
					// Verify first coffee item has Computed attributes updated.
					resource.TestCheckResourceAttr("hwmux_device.test", "is_wstk", "false"),
					resource.TestCheckResourceAttr("hwmux_device.test", "metadata", "{}"),
					resource.TestCheckResourceAttr("hwmux_device.test", "location_metadata", "{}"),
					resource.TestCheckResourceAttr("hwmux_device.test", "source", "TERRAFORM"),
					resource.TestCheckResourceAttr("hwmux_device.test", "socketed_chip", "BRD1019A002"),
				),
			},
			// (Delete testing automatically occurs in TestCase)
		},
	})
}

func TestAccDeviceResourceNoUri(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "hwmux_device" "test" {
	sn_or_name = "test_device"
    part = "Part_no_0"
    room = "Room_0"
    permission_groups = ["Staff users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_device.test", "sn_or_name", "test_device"),
					resource.TestCheckResourceAttr("hwmux_device.test", "part", "Part_no_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "room", "Room_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_device.test", "permission_groups.0", "Staff users"),
					// Verify the device item has Computed attributes filled.
					resource.TestCheckResourceAttr("hwmux_device.test", "is_wstk", "false"),
					resource.TestCheckResourceAttr("hwmux_device.test", "online", "true"),
					resource.TestCheckResourceAttr("hwmux_device.test", "metadata", "{}"),
					resource.TestCheckResourceAttr("hwmux_device.test", "location_metadata", "{}"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("hwmux_device.test", "id"),
					resource.TestCheckResourceAttrSet("hwmux_device.test", "last_updated"),
					resource.TestCheckResourceAttr("hwmux_device.test", "source", "TERRAFORM"),
					// Verify if optional field is set to default
					resource.TestCheckResourceAttr("hwmux_device.test", "socketed_chip", ""),
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
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDeviceResourceNoSnOrName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "hwmux_device" "test" {
    part = "Part_no_0"
    room = "Room_0"
    wstk_part = "Part_no_0"
    permission_groups = ["Staff users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_device.test", "part", "Part_no_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "wstk_part", "Part_no_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "room", "Room_0"),
					resource.TestCheckResourceAttr("hwmux_device.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_device.test", "permission_groups.0", "Staff users"),
					// Verify the device item has Computed attributes filled.
					resource.TestCheckResourceAttr("hwmux_device.test", "is_wstk", "false"),
					resource.TestCheckResourceAttr("hwmux_device.test", "online", "true"),
					resource.TestCheckResourceAttr("hwmux_device.test", "metadata", "{}"),
					resource.TestCheckResourceAttr("hwmux_device.test", "location_metadata", "{}"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("hwmux_device.test", "id"),
					resource.TestCheckResourceAttrSet("hwmux_device.test", "last_updated"),
					resource.TestCheckResourceAttr("hwmux_device.test", "source", "TERRAFORM"),
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
			// Delete testing automatically occurs in TestCase
		},
	})
}
