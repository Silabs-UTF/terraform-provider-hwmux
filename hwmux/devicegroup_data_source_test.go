package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDeviceGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "hwmux_device_group" "test" {id = 1}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the first coffee to ensure all attributes are set
					resource.TestCheckResourceAttr("data.hwmux_device_group.test", "name", "group0"),
					resource.TestCheckResourceAttr("data.hwmux_device_group.test", "id", "1"),
					// ensure dynamic attributes are populated
					resource.TestCheckResourceAttrSet("data.hwmux_device_group.test", "enable_ahs"),
					resource.TestCheckResourceAttrSet("data.hwmux_device_group.test", "enable_ahs_actions"),
					resource.TestCheckResourceAttrSet("data.hwmux_device_group.test", "metadata"),
					resource.TestCheckResourceAttrSet("data.hwmux_device_group.test", "devices.0.id"),
				),
			},
		},
	})
}
