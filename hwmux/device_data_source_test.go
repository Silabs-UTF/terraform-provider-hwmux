package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDeviceDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "hwmux_device" "test" {id = 1}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the first coffee to ensure all attributes are set
					resource.TestCheckResourceAttr("data.hwmux_device.test", "sn_or_name", "sn0"),
					resource.TestCheckResourceAttr("data.hwmux_device.test", "id", "1"),
					resource.TestCheckResourceAttr("data.hwmux_device.test", "is_wstk", "false"),
					resource.TestCheckResourceAttr("data.hwmux_device.test", "wstk_part", ""),
					resource.TestCheckResourceAttr("data.hwmux_device.test", "online", "true"),
					resource.TestCheckResourceAttr("data.hwmux_device.test", "part", "Part_no_0"),
					resource.TestCheckResourceAttr("data.hwmux_device.test", "uri", "0"),
				),
			},
		},
	})
}
