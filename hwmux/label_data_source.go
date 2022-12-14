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
var _ datasource.DataSource = &LabelDataSource{}

func NewLabelDataSource() datasource.DataSource {
	return &LabelDataSource{}
}

type LabelDataSource struct {
	client *hwmux.APIClient
}

// labelDataSourceModel maps the data source schema data.
type LabelDataSourceModel struct {
	ID           types.Int64              `tfsdk:"id"`
	Name         types.String             `tfsdk:"name"`
	DeviceGroups []nestedDeviceGroupModel `tfsdk:"device_groups"`
	Metadata     types.String             `tfsdk:"metadata"`
}

type nestedDeviceGroupModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}


func (d *LabelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (d *LabelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Label data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Label identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Label name. Must be unique.",
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "The metadata of the Label.",
				Computed: true,
			},
			"device_groups": schema.ListNestedAttribute{
				MarkdownDescription: "The Device Groups that belong to the Label",
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Device Group ID.",
							Computed: true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Device Group name.",
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *LabelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LabelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LabelDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	label, _, err := GetLabel(d.client, &resp.Diagnostics, int32(data.ID.ValueInt64()))
	if err != nil {
		return
	}

	// Map response body to model
	data.ID = types.Int64Value(int64(label.GetId()))
	data.Name = types.StringValue(label.GetName())

	err = MarshalMetadataSetError(label.GetMetadata(), &resp.Diagnostics, "label", &data.Metadata)
	if err != nil {
		return
	}

	data.DeviceGroups = make([]nestedDeviceGroupModel, len(label.GetDeviceGroups()))
	for i, deviceGroup := range label.GetDeviceGroups() {
		fullDeviceGroup, _, err := GetDeviceGroup(d.client, &resp.Diagnostics, deviceGroup)
		if err != nil {
			return
		}

		data.DeviceGroups[i] = nestedDeviceGroupModel{
			ID:   types.Int64Value(int64(deviceGroup)),
			Name: types.StringValue(fullDeviceGroup.GetName()),
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
