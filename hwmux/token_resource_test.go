package hwmux

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTokenResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: providerConfig + `
resource "hwmux_token" "test" {
	user_id     = "dev1"
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("hwmux_token.test", "user_id", "dev1"),
					resource.TestCheckResourceAttrSet("hwmux_token.test", "token"),
                    resource.TestCheckResourceAttrSet("hwmux_token.test", "id"),
					resource.TestCheckResourceAttrSet("hwmux_token.test", "date_created"),
                    resource.TestCheckResourceAttrSet("hwmux_token.test", "last_updated"),
                ),
            },
            // Update and Read testing
            {
                Config: providerConfig + `
resource "hwmux_token" "test" {
	user_id     = "dev2"
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hwmux_token.test", "user_id", "dev2"),
					resource.TestCheckResourceAttrSet("hwmux_token.test", "token"),
                    resource.TestCheckResourceAttrSet("hwmux_token.test", "id"),
					resource.TestCheckResourceAttrSet("hwmux_token.test", "date_created"),
                    resource.TestCheckResourceAttrSet("hwmux_token.test", "last_updated"),
                ),
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}
