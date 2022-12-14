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
var _ datasource.DataSource = &DeviceDataSource{}

func NewDeviceDataSource() datasource.DataSource {
	return &DeviceDataSource{}
}

type DeviceDataSource struct {
	client *hwmux.APIClient
}

// deviceDataSourceModel maps the data source schema data.
type DeviceDataSourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Sn_or_name types.String `tfsdk:"sn_or_name"`
	Is_wstk    types.Bool   `tfsdk:"is_wstk"`
	Uri        types.String `tfsdk:"uri"`
	Online     types.Bool   `tfsdk:"online"`
	Metadata   types.String `tfsdk:"metadata"`
	Part       types.String `tfsdk:"part"`
}

func (d *DeviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (d *DeviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Device data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Device identifier",
				Required:            true,
			},
			"sn_or_name": schema.StringAttribute{
				MarkdownDescription: "Device name. Must be unique.",
				Computed:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "The URI or IP address of the device.",
				Computed:            true,
			},
			"part": schema.StringAttribute{
				MarkdownDescription: "The part number of the device.",
				Computed:            true,
			},
			"is_wstk": schema.BoolAttribute{
				MarkdownDescription: "If the device is a WSTK.",
				Computed: true,
			},
			"online": schema.BoolAttribute{
				MarkdownDescription: "If the device is online.",
				Computed: true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "The metadata of the device.",
				Computed: true,
			},
		},
	}
}

func (d *DeviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeviceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	device, _, err := GetDevice(d.client, &resp.Diagnostics, int32(data.ID.ValueInt64()))
	if err != nil {
		return
	}

	data.ID = types.Int64Value(int64(device.GetId()))
	data.Sn_or_name = types.StringValue(device.GetSnOrName())
	data.Is_wstk = types.BoolValue(device.GetIsWstk())
	data.Uri = types.StringValue(device.GetUri())
	data.Online = types.BoolValue(device.GetOnline())

	err = MarshalMetadataSetError(device.GetMetadata(), &resp.Diagnostics, "device", &data.Metadata)
	if err != nil {
		return
	}

	data.Part = types.StringValue(device.Part.GetPartNo())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
