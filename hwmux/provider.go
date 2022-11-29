package hwmux

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"stash.silabs.com/iot_infra_sw/hwmux-client-golang"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &hwmuxProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &hwmuxProvider{}
}

// hwmuxProvider is the provider implementation.
type hwmuxProvider struct{}

// Metadata returns the provider type name.
func (p *hwmuxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hwmux"
}

// hwmuxProviderModel maps provider schema data to a Go type.
type hwmuxProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

// GetSchema defines the provider-level schema for configuration data.
func (p *hwmuxProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				Type:     types.StringType,
				Optional: true,
			},
			"token": {
				Type:     types.StringType,
				Optional: true,
				Sensitive: true,
			},
		},
	}, nil
}

// Configure prepares a hwmux API client for data sources and resources.
func (p *hwmuxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Hwmux client")

	// Retrieve provider data from configuration
	var config hwmuxProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown hwmux API Host",
			"The provider cannot create the hwmux API client as there is an unknown configuration value for the hwmux API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the hwmux_HOST environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown hwmux API token",
			"The provider cannot create the hwmux API client as there is an unknown configuration value for the hwmux API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the hwmux_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("HWMUX_HOST")
	token := os.Getenv("HWMUX_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing hwmux API Host",
			"The provider cannot create the hwmux API client as there is a missing or empty value for the hwmux API host. "+
				"Set the host value in the configuration or use the hwmux_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing hwmux API Token",
			"The provider cannot create the hwmux API client as there is a missing or empty value for the hwmux API token. "+
				"Set the token value in the configuration or use the hwmux_token environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "hwmux_host", host)
    ctx = tflog.SetField(ctx, "hwmux_token", token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "hwmux_token")

    tflog.Debug(ctx, "Creating Hwmux client")

	// Create a new hwmux client using the configuration values
	clientConfig := hwmux.NewConfiguration()
	clientConfig.AddDefaultHeader("Authorization", "Token " + token)
	clientConfig.Servers = hwmux.ServerConfigurations{hwmux.ServerConfiguration{URL: host}}
	client := hwmux.NewAPIClient(clientConfig)

	// Make the hwmux client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Hwmux client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *hwmuxProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDeviceDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *hwmuxProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
