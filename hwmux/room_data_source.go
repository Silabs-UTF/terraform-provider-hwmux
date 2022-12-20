package hwmux

import (
	"context"
	"fmt"

	"github.com/Silabs-UTF/hwmux-client-golang"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &RoomDataSource{}

func NewRoomDataSource() datasource.DataSource {
	return &RoomDataSource{}
}

type RoomDataSource struct {
	client *hwmux.APIClient
}

// roomDataSourceModel maps the data source schema data.
type RoomDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Site     types.String `tfsdk:"site"`
	Metadata types.String `tfsdk:"metadata"`
}


func (d *RoomDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_room"
}

func (d *RoomDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Room data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Part identifier. Always equals the name. Set to satisfy terraform restrictions.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Room name.",
				Required:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Site.",
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "The metadata of the Room.",
				Computed: true,
			},
		},
	}
}

func (d *RoomDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RoomDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoomDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Map response body to model
	room, _, err := GetRoom(d.client, &resp.Diagnostics, data.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to model
	data.ID = types.StringValue(room.GetName())
	data.Name = types.StringValue(room.GetName())
	data.Site = types.StringValue(room.GetSite())

	err = MarshalMetadataSetError(room.GetMetadata(), &resp.Diagnostics, "room", &data.Metadata)
	if err != nil {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
