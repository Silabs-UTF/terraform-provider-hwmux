package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)


func TestAccLabelDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Read testing
            {
                Config: providerConfig + `data "hwmux_label" "test" {id = 1}`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Verify the first coffee to ensure all attributes are set
                    resource.TestCheckResourceAttr("data.hwmux_label.test", "name", "label0"),
                    resource.TestCheckResourceAttr("data.hwmux_label.test", "id", "1"),
                    resource.TestCheckResourceAttr("data.hwmux_label.test", "device_groups.#", "8"),
					resource.TestCheckResourceAttrSet("data.hwmux_label.test", "device_groups.0.id"),
                    resource.TestCheckResourceAttrSet("data.hwmux_label.test", "device_groups.0.name"),
                    resource.TestCheckResourceAttrSet("data.hwmux_label.test", "metadata"),
                ),
            },
        },
	})
}

