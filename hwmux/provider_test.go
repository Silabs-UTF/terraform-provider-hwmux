package hwmux

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "hwmux" {
  host  = "http://localhost"
  token = "8164a97ba1324698e838146781a7365fb969edc4"
}
`
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hwmux": providerserver.NewProtocol6WithError(New("test")()),
}
