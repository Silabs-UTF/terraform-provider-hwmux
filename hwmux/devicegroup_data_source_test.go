package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var deviceGroupDataSourceTfName string = "data.hwmux_device_group.test"

func TestAccDeviceGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "hwmux_device_group" "test" {id = 1}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the first coffee to ensure all attributes are set
					resource.TestCheckResourceAttr(deviceGroupDataSourceTfName, "name", "group0"),
					resource.TestCheckResourceAttr(deviceGroupDataSourceTfName, "id", "1"),
					// ensure dynamic attributes are populated
					resource.TestCheckResourceAttrSet(deviceGroupDataSourceTfName, "enable_ahs"),
					resource.TestCheckResourceAttrSet(deviceGroupDataSourceTfName, "enable_ahs_actions"),
					resource.TestCheckResourceAttrSet(deviceGroupDataSourceTfName, "enable_ahs_cas"),
					resource.TestCheckResourceAttrSet(deviceGroupDataSourceTfName, "metadata"),
					resource.TestCheckResourceAttrSet(deviceGroupDataSourceTfName, "devices.0.id"),
				),
			},
		},
	})
}
