package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)


func TestAccPartDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Read testing
            {
                Config: providerConfig + `data "hwmux_part" "test" {part_no = "Part_no_0"}`,
                Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hwmux_part.test", "id", "Part_no_0"),
                    resource.TestCheckResourceAttr("data.hwmux_part.test", "part_no", "Part_no_0"),
					resource.TestCheckResourceAttr("data.hwmux_part.test", "board_no", "Part_no_0"),
					resource.TestCheckResourceAttr("data.hwmux_part.test", "chip_no", "chip0"),
					resource.TestCheckResourceAttr("data.hwmux_part.test", "variant", "A"),
					resource.TestCheckResourceAttr("data.hwmux_part.test", "revision", "A00"),
					resource.TestCheckResourceAttr("data.hwmux_part.test", "part_family.name", "PartFamily_0"),
					resource.TestCheckResourceAttrSet("data.hwmux_part.test", "metadata"),
                ),
            },
        },
	})
}
