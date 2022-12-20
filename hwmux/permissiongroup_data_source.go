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
var _ datasource.DataSource = &PermissionGroupDataSource{}

func NewPermissionGroupDataSource() datasource.DataSource {
	return &PermissionGroupDataSource{}
}

type PermissionGroupDataSource struct {
	client *hwmux.APIClient
}

// permissionGroupDataSourceModel maps the data source schema data.
type PermissionGroupDataSourceModel struct {
	Name        types.String   `tfsdk:"name"`
	ID          types.Int64    `tfsdk:"id"`
	Permissions []types.String `tfsdk:"permissions"`
}

func (d *PermissionGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_group"
}

func (d *PermissionGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Permission group data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Permission group name",
				Required:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "Permission group identifier",
				Computed:            true,
			},
			"permissions": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Permissions assigned to this permission group",
				Computed:            true,
			},
		},
	}
}

func (d *PermissionGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PermissionGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PermissionGroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	permissionGroup, _, err := GetPermissionGroup(d.client, &resp.Diagnostics, data.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to model
	data.Name = types.StringValue(permissionGroup.GetName())
	data.ID = types.Int64Value(int64(permissionGroup.GetId()))

	data.Permissions = make([]types.String, len(permissionGroup.GetPermissions()))
	for i, permission := range permissionGroup.GetPermissions() {
		data.Permissions[i] = types.StringValue(permission)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
