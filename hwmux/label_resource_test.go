package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLabelResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "hwmux_label" "test" {
	name     = "test_label_tf"
	device_groups = [1, 2]
	permission_groups = ["All users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_label.test", "name", "test_label_tf"),
					resource.TestCheckResourceAttr("hwmux_label.test", "device_groups.#", "2"),
					resource.TestCheckResourceAttr("hwmux_label.test", "device_groups.0", "1"),
					resource.TestCheckResourceAttr("hwmux_label.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_label.test", "permission_groups.0", "All users"),
					// Verify the label item has Computed attributes filled.
					resource.TestCheckResourceAttr("hwmux_label.test", "metadata", "{}"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("hwmux_label.test", "id"),
					resource.TestCheckResourceAttrSet("hwmux_label.test", "last_updated"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hwmux_label.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "hwmux_label" "test" {
    name     = "test_label_tf"
    device_groups = [1]
    permission_groups = ["Staff users"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_label.test", "name", "test_label_tf"),
					resource.TestCheckResourceAttr("hwmux_label.test", "device_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_label.test", "device_groups.0", "1"),
					resource.TestCheckResourceAttr("hwmux_label.test", "permission_groups.#", "1"),
					resource.TestCheckResourceAttr("hwmux_label.test", "permission_groups.0", "Staff users"),
					resource.TestCheckResourceAttr("hwmux_label.test", "metadata", "{}"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
