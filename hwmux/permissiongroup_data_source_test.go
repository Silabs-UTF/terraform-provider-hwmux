package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)


func TestAccPermissionGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Read testing
            {
                Config: providerConfig + `data "hwmux_permission_group" "test" {name = "All users"}`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.hwmux_permission_group.test", "name", "All users"),
                    resource.TestCheckResourceAttrSet("data.hwmux_permission_group.test", "permissions.#"),
                    resource.TestCheckResourceAttrSet("data.hwmux_permission_group.test", "id"),
                ),
            },
        },
	})
}
