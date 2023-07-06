package hwmux

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "hwmux" {
  host  = "http://localhost"
  token = "a86690a0a1e93c139db8858e039104b30ba38d2c"
}
`
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hwmux": providerserver.NewProtocol6WithError(New("test")()),
}
