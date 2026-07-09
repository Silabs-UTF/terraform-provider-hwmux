package hwmux

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "hwmux" {
  host  = "http://localhost"
  token = "2f2d6b0556dd411b17637e5da6f36878842dcbc3"
}
`
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hwmux": providerserver.NewProtocol6WithError(New("test")()),
}
