package hwmux

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"stash.silabs.com/iot_infra_sw/hwmux-client-golang"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &DeviceGroupDataSource{}

func NewDeviceGroupDataSource() datasource.DataSource {
	return &DeviceGroupDataSource{}
}

type DeviceGroupDataSource struct {
	client *hwmux.APIClient
}

// deviceGroupDataSourceModel maps the data source schema data.
type DeviceGroupDataSourceModel struct {
	ID       types.Int64         `tfsdk:"id"`
	Name     types.String        `tfsdk:"name"`
	Devices  []nestedDeviceModel `tfsdk:"devices"`
	Metadata types.String        `tfsdk:"metadata"`
}

type nestedDeviceModel struct {
	ID types.Int64 `tfsdk:"id"`
}


func (d *DeviceGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_group"
}

func (d *DeviceGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "DeviceGroup data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Device Group identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Device Group name. Must be unique.",
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "The metadata of the Device Group.",
				Computed: true,
			},
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "The devices that belong to the Device Group",
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Device ID.",
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *DeviceGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hwmux.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hwmux.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DeviceGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeviceGroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deviceGroup, _, err := GetDeviceGroup(d.client, &resp.Diagnostics, int32(data.ID.ValueInt64()))
	if err != nil {
		return
	}

	// Map response body to model
	data.ID = types.Int64Value(int64(deviceGroup.GetId()))
	data.Name = types.StringValue(deviceGroup.GetName())

	err = MarshalMetadataSetError(deviceGroup.GetMetadata(), &resp.Diagnostics, "deviceGroup", &data.Metadata)
	if err != nil {
		return
	}

	data.Devices = make([]nestedDeviceModel, len(deviceGroup.GetDevices()))
	for i, device := range deviceGroup.GetDevices() {
		data.Devices[i] = nestedDeviceModel{ID: types.Int64Value(int64(device.GetId()))}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
