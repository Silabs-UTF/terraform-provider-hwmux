package hwmux

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.Int64Type,
				Required: true,
			},
			"sn_or_name": {
				Type:     types.StringType,
				Computed: true,
			},
			"is_wstk": {
				Type:     types.BoolType,
				Computed: true,
			},
			"uri": {
				Type:     types.StringType,
				Computed: true,
			},
			"online": {
				Type:     types.BoolType,
				Computed: true,
			},
			"metadata": {
				Type: types.StringType,
				Computed: true,
			},
			"part": {
				Computed: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"part_no": {
						Type: types.StringType,
						Computed: true,
					},
				}),
			},
		},
	}, nil
}

// deviceDataSourceModel maps the data source schema data.
type deviceDataSourceModel struct {
	ID         types.Int64     `tfsdk:"id"`
	Sn_or_name types.String    `tfsdk:"sn_or_name"`
	Is_wstk    types.Bool      `tfsdk:"is_wstk"`
	Uri        types.String    `tfsdk:"uri"`
	Online     types.Bool      `tfsdk:"online"`
	Metadata   types.String	   `tfsdk:"metadata"`
	Part       devicePartModel `tfsdk:"part"`
}

// devicePartModel maps device ingredients data
type devicePartModel struct {
	Part_no types.String `tfsdk:"part_no"`
}

// Read refreshes the Terraform state with the latest data.
func (d *deviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceDataSourceModel
	var id int32
	
	diags := req.Config.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)

	device, api_response, err := d.client.DevicesApi.DevicesRetrieve(context.Background(), id).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Device",
			err.Error(),
		)
		return
	}
	if api_response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Api response code: %d", api_response.StatusCode),
			err.Error(),
		)
		return
	}

	// Map response body to model
	state.ID = types.Int64Value(int64(device.Id))
	state.Sn_or_name = types.StringValue(device.GetSnOrName())
	state.Is_wstk = types.BoolValue(device.GetIsWstk())
	state.Uri = types.StringValue(device.GetUri())
	state.Online = types.BoolValue(device.GetOnline())

	metadataJson, err := json.Marshal(device.GetMetadata())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get device metadata",
			err.Error(),
		)
		return
	}
	state.Metadata = types.StringValue(string(metadataJson))

	state.Part = devicePartModel{
		Part_no: types.StringValue(device.Part.GetPartNo()),
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
