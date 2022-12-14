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
var _ datasource.DataSource = &PartDataSource{}

func NewPartDataSource() datasource.DataSource {
	return &PartDataSource{}
}

type PartDataSource struct {
	client *hwmux.APIClient
}

// partDataSourceModel maps the data source schema data.
type PartDataSourceModel struct {
	ID          types.String          `tfsdk:"id"`
	Part_no     types.String          `tfsdk:"part_no"`
	Board_no    types.String          `tfsdk:"board_no"`
	Chip_no     types.String          `tfsdk:"chip_no"`
	Variant     types.String          `tfsdk:"variant"`
	Revision    types.String          `tfsdk:"revision"`
	Part_family *nestedPartFamilyModel `tfsdk:"part_family"`
	Metadata    types.String          `tfsdk:"metadata"`
}

type nestedPartFamilyModel struct {
	Name types.String `tfsdk:"name"`
}


func (d *PartDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_part"
}

func (d *PartDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Part data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Part identifier. Always equals the part_no. Set to satisfy terraform restrictions.",
				Computed:            true,
			},
			"part_no": schema.StringAttribute{
				MarkdownDescription: "Part number.",
				Required:            true,
			},
			"board_no": schema.StringAttribute{
				MarkdownDescription: "Board number.",
				Computed:            true,
			},
			"chip_no": schema.StringAttribute{
				MarkdownDescription: "Chip number.",
				Computed:            true,
			},
			"revision": schema.StringAttribute{
				MarkdownDescription: "Part revision.",
				Computed:            true,
			},
			"variant": schema.StringAttribute{
				MarkdownDescription: "Part variant.",
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "The metadata of the Part.",
				Computed: true,
			},
			"part_family": schema.SingleNestedAttribute{
				MarkdownDescription: "The Part Family.",
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "The part family name.",
						Computed: true,
					},
				},
			},
		},
	}
}

func (d *PartDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PartDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PartDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	part, _, err := GetPart(d.client, &resp.Diagnostics, data.Part_no.ValueString())
	if err != nil {
		return
	}

	// Map response body to model
	data.ID = types.StringValue(part.GetPartNo())
	data.Part_no = types.StringValue(part.GetPartNo())
	data.Board_no = types.StringValue(part.GetBoardNo())
	data.Chip_no = types.StringValue(part.GetChipNo())
	data.Variant = types.StringValue(part.GetVariant())
	data.Revision = types.StringValue(part.GetRevision())

	err = MarshalMetadataSetError(part.GetMetadata(), &resp.Diagnostics, "part", &data.Metadata)
	if err != nil {
		return
	}

	data.Part_family = &nestedPartFamilyModel{Name: types.StringValue(part.PartFamily.GetName())}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
