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
	_ datasource.DataSource              = &deviceGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &deviceGroupDataSource{}
)

// NewDeviceGroupDataSource is a helper function to simplify the provider implementation.
func NewDeviceGroupDataSource() datasource.DataSource {
	return &deviceGroupDataSource{}
}

// deviceGroupDataSource is the data source implementation.
type deviceGroupDataSource struct {
	client *hwmux.APIClient
}

func (d *deviceGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*hwmux.APIClient)
}

// Metadata returns the data source type name.
func (d *deviceGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deviceGroup"
}

// GetSchema defines the schema for the data source.
func (d *deviceGroupDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Fetches a deviceGroup from hwmux.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "The ID of the deviceGroup.",
				Type:        types.Int64Type,
				Required:    true,
			},
			"name": {
				Description: "The name of the deviceGroup. Must be unique.",
				Type:        types.StringType,
				Computed:    true,
			},
			"metadata": {
				Description: "The metadata of the deviceGroup.",
				Type:        types.StringType,
				Computed:    true,
			},
			"devices": {
				Description: "The devices that belong to the deviceGroup.",
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "The ID of the device.",
						Type:        types.Int64Type,
						Computed:    true,
					},
				}),
			},
		},
	}, nil
}

// deviceGroupDataSourceModel maps the data source schema data.
type deviceGroupDataSourceModel struct {
	ID       types.Int64         `tfsdk:"id"`
	Name     types.String        `tfsdk:"name"`
	Devices  []nestedDeviceModel `tfsdk:"devices"`
	Metadata types.String        `tfsdk:"metadata"`
}

type nestedDeviceModel struct {
	ID types.Int64 `tfsdk:"id"`
}

// Read refreshes the Terraform state with the latest data.
func (d *deviceGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceGroupDataSourceModel
	var id int64

	diags := req.Config.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)

	deviceGroup, _, err := GetDeviceGroup(d.client, &resp.Diagnostics, int32(id))
	if err != nil {
		return
	}

	// Map response body to model
	state.ID = types.Int64Value(int64(deviceGroup.GetId()))
	state.Name = types.StringValue(deviceGroup.GetName())

	err = MarshalMetadataSetError(deviceGroup.GetMetadata(), &resp.Diagnostics, "deviceGroup", &state.Metadata)
	if err != nil {
		return
	}

	for _, device := range deviceGroup.GetDevices() {
		state.Devices = append(state.Devices, nestedDeviceModel{ID: types.Int64Value(int64(device.GetId()))})
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
