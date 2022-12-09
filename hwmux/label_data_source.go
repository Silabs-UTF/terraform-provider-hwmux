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
	_ datasource.DataSource              = &labelDataSource{}
	_ datasource.DataSourceWithConfigure = &labelDataSource{}
)

// NewLabelDataSource is a helper function to simplify the provider implementation.
func NewLabelDataSource() datasource.DataSource {
	return &labelDataSource{}
}

// labelDataSource is the data source implementation.
type labelDataSource struct {
	client *hwmux.APIClient
}

func (d *labelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*hwmux.APIClient)
}

// Metadata returns the data source type name.
func (d *labelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

// GetSchema defines the schema for the data source.
func (d *labelDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Fetches a label from hwmux.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "The ID of the label.",
				Type:        types.Int64Type,
				Required:    true,
			},
			"name": {
				Description: "The name of the label. Must be unique.",
				Type:        types.StringType,
				Computed:    true,
			},
			"metadata": {
				Description: "The metadata of the label.",
				Type:        types.StringType,
				Computed:    true,
			},
			"device_groups": {
				Description: "The deviceGroups that belong to the label.",
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Description: "The ID of the deviceGroup.",
						Type:        types.Int64Type,
						Computed:    true,
					},
					"name": {
						Description: "The name of the deviceGroup.",
						Type:        types.StringType,
						Computed:    true,
					},
				}),
			},
		},
	}, nil
}

// labelDataSourceModel maps the data source schema data.
type labelDataSourceModel struct {
	ID           types.Int64              `tfsdk:"id"`
	Name         types.String             `tfsdk:"name"`
	DeviceGroups []nestedDeviceGroupModel `tfsdk:"device_groups"`
	Metadata     types.String             `tfsdk:"metadata"`
}

type nestedDeviceGroupModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Read refreshes the Terraform state with the latest data.
func (d *labelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state labelDataSourceModel
	var id int64

	diags := req.Config.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)

	label, _, err := GetLabel(d.client, &resp.Diagnostics, int32(id))
	if err != nil {
		return
	}

	// Map response body to model
	state.ID = types.Int64Value(int64(label.GetId()))
	state.Name = types.StringValue(label.GetName())

	err = MarshalMetadataSetError(label.GetMetadata(), &resp.Diagnostics, "label", &state.Metadata)
	if err != nil {
		return
	}

	state.DeviceGroups = make([]nestedDeviceGroupModel, len(label.GetDeviceGroups()))
	for i, deviceGroup := range label.GetDeviceGroups() {
		fullDeviceGroup, _, err := GetDeviceGroup(d.client, &resp.Diagnostics, deviceGroup)
		if err != nil {
			return
		}

		state.DeviceGroups[i] = nestedDeviceGroupModel{
			ID:   types.Int64Value(int64(deviceGroup)),
			Name: types.StringValue(fullDeviceGroup.GetName()),
		}
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
