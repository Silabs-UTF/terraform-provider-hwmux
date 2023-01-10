package hwmux

import (
	"context"
	"os"

	"github.com/Silabs-UTF/hwmux-client-golang"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure HwmuxProvider satisfies various provider interfaces.
var _ provider.Provider = &HwmuxProvider{}

// HwmuxProvider defines the provider implementation.
type HwmuxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HwmuxProviderModel describes the provider data model.
type HwmuxProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

func (p *HwmuxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hwmux"
	resp.Version = p.version
}

func (p *HwmuxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interact with Hwmux.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "URI to Hwmux API. May also be provided via HWMUX_HOST environment variable. No trailing slash required.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The Hwmux API token. May also be provided via HWMUX_TOKEN environment variable.",
				Optional: true,
				Sensitive: true,
			},
		},
	}
}

func (p *HwmuxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data HwmuxProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown hwmux API Host",
			"The provider cannot create the hwmux API client as there is an unknown configuration value for the hwmux API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the hwmux_HOST environment variable.",
		)
	}

	if data.Token.IsUnknown() {
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

	if !data.Host.IsNull() {
		host = data.Host.ValueString()
	}

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
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
	clientConfig.AddDefaultHeader("Authorization", "Token "+token)
	clientConfig.Servers = hwmux.ServerConfigurations{hwmux.ServerConfiguration{URL: host}}
	client := hwmux.NewAPIClient(clientConfig)
	
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *HwmuxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDeviceResource,
		NewDeviceGroupResource,
		NewLabelResource,
		NewPermissionGroupResource,
		NewUserResource,
	}
}

func (p *HwmuxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDeviceDataSource,
		NewDeviceGroupDataSource,
		NewLabelDataSource,
		NewPermissionGroupDataSource,
		NewPartDataSource,
		NewRoomDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HwmuxProvider{
			version: version,
		}
	}
}
