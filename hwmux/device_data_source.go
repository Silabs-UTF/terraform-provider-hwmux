package hwmux

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"stash.silabs.com/iot_infra_sw/hwmux-client-golang"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &deviceDataSource{}
	_ datasource.DataSourceWithConfigure = &deviceDataSource{}
)

// NewDeviceDataSource is a helper function to simplify the provider implementation.
func NewDeviceDataSource() datasource.DataSource {
	return &deviceDataSource{}
}

// deviceDataSource is the data source implementation.
type deviceDataSource struct {
	client *hwmux.APIClient
}

func (d *deviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*hwmux.APIClient)
}

// Metadata returns the data source type name.
func (d *deviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

// GetSchema defines the schema for the data source.
func (d *deviceDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Fetches a device from hwmux.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "The ID of the device.",
				Type:        types.Int64Type,
				Required:    true,
			},
			"sn_or_name": {
				Description: "The name of the device. Must be unique.",
				Type:        types.StringType,
				Computed:    true,
			},
			"is_wstk": {
				Description: "Whether the device is a WSTK.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"uri": {
				Description: "The URI or IP address of the device.",
				Type:        types.StringType,
				Computed:    true,
			},
			"online": {
				Description: "Whether the device is online.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"metadata": {
				Description: "The metadata of the device.",
				Type:        types.StringType,
				Computed:    true,
			},
			"part": {
				Description: "The part number of the device.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

// deviceDataSourceModel maps the data source schema data.
type deviceDataSourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Sn_or_name types.String `tfsdk:"sn_or_name"`
	Is_wstk    types.Bool   `tfsdk:"is_wstk"`
	Uri        types.String `tfsdk:"uri"`
	Online     types.Bool   `tfsdk:"online"`
	Metadata   types.String `tfsdk:"metadata"`
	Part       types.String `tfsdk:"part"`
}

// Read refreshes the Terraform state with the latest data.
func (d *deviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceDataSourceModel
	var id int64

	diags := req.Config.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)

	device, _, err := GetDevice(d.client, &resp.Diagnostics, int32(id))
	if err != nil {
		return
	}

	// Map response body to model
	state.ID = types.Int64Value(int64(device.GetId()))
	state.Sn_or_name = types.StringValue(device.GetSnOrName())
	state.Is_wstk = types.BoolValue(device.GetIsWstk())
	state.Uri = types.StringValue(device.GetUri())
	state.Online = types.BoolValue(device.GetOnline())

	err = MarshalMetadataSetError(device.GetMetadata(), &resp.Diagnostics, "device", &state.Metadata)
	if err != nil {
		return
	}

	state.Part = types.StringValue(device.Part.GetPartNo())

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
