package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRoomDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `data "hwmux_room" "test" {name = "Room_0"}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hwmux_room.test", "id", "Room_0"),
					resource.TestCheckResourceAttr("data.hwmux_room.test", "name", "Room_0"),
					resource.TestCheckResourceAttr("data.hwmux_room.test", "site", "Site_0"),
					resource.TestCheckResourceAttrSet("data.hwmux_room.test", "metadata"),
				),
			},
		},
	})
}
