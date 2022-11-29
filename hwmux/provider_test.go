package hwmux

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)


const (
	providerConfig = `
provider "hwmux" {
  host  = "http://localhost"
  token = "cd48869f98f4eca6eb3bda542f5ac808e4beabc4"
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
        "hwmux": providerserver.NewProtocol6WithError(New()),
    }
)

