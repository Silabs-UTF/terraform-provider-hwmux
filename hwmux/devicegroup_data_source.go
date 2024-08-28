package hwmux

import (
	"context"
	"fmt"

	"github.com/Silabs-UTF/hwmux-client-golang/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ID                 types.Int64         `tfsdk:"id"`
	Name               types.String        `tfsdk:"name"`
	Devices            []nestedDeviceModel `tfsdk:"devices"`
	Enable_ahs         types.Bool          `tfsdk:"enable_ahs"`
	Enable_ahs_cas     types.Bool          `tfsdk:"enable_ahs_cas"`
	Enable_ahs_actions types.Bool          `tfsdk:"enable_ahs_actions"`
	Metadata           types.String        `tfsdk:"metadata"`
	Source             types.String        `tfsdk:"source"`
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
				Computed:            true,
			},
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "The devices that belong to the Device Group",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Device ID.",
							Computed:            true,
						},
					},
				},
			},
			"enable_ahs": schema.BoolAttribute{
				MarkdownDescription: "Enable the Automated Health Service",
				Computed:            true,
			},
			"enable_ahs_actions": schema.BoolAttribute{
				MarkdownDescription: "Allow the Automated Health Service to take DeviceGroups offline when they are unhealthy.",
				Computed:            true,
			},
			"enable_ahs_cas": schema.BoolAttribute{
				MarkdownDescription: "Enable the Automated Health Service to take Corrective Actions.",
				Computed:            true,
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The source where the device group was created.",
				Computed:            true,
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
	data.Enable_ahs = types.BoolValue(deviceGroup.GetEnableAhs())
	data.Enable_ahs_actions = types.BoolValue(deviceGroup.GetEnableAhsActions())
	data.Enable_ahs_cas = types.BoolValue(deviceGroup.GetEnableAhsCas())
	data.Source = types.StringValue(string(deviceGroup.GetSource()))

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
