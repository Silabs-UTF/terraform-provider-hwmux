package hwmux

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Silabs-UTF/hwmux-client-golang"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &PermissionGroupResource{}

func NewPermissionGroupResource() resource.Resource {
	return &PermissionGroupResource{}
}

// PermissionGroupResource defines the resource implementation.
type PermissionGroupResource struct {
	client *hwmux.APIClient
}

// PermissionGroupResourceModel describes the resource data model.
type PermissionGroupResourceModel struct {
	ID			types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Permissions types.Set `tfsdk:"permissions"`
	LastUpdated types.String   `tfsdk:"last_updated"`
}

func (r *PermissionGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_group"
}

func (r *PermissionGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "PermissionGroup resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Permission Group identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Permission Group name",
			},
			"permissions": schema.SetAttribute{
				MarkdownDescription: "The permissions that this permission group has.",
				ElementType: types.StringType,
				Computed: true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the resource.",
				Computed: true,
			},
		},
	}
}

func (r *PermissionGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hwmux.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *hwmux.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PermissionGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *PermissionGroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	permissionGroupSerializer, err := createPermissionGroupFromPlan(data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create permissionGroup API request based on plan", err.Error(),
		)
		return
	}

	// create new permissionGroup
	permissionGroupSerializer, httpRes, err := r.client.PermissionsApi.PermissionsGroupsCreate(context.Background()).PermissionGroup(*permissionGroupSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating permissionGroup",
			"Could not create permissionGroup, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	// set model based on response
	err = updatePermissionGroupModelFromResponse(permissionGroupSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the permissionGroup model failed", err.Error(),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *PermissionGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed permissionGroup value from hwmux
	permissionGroup, _, err := GetPermissionGroup(r.client, &resp.Diagnostics, data.Name.ValueString())
	if err != nil {
		return
	}

	// Map response body to model
	err = updatePermissionGroupModelFromResponse(permissionGroup, data, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *PermissionGroupResourceModel
	var state *PermissionGroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	permissionGroupSerializer := hwmux.NewPermissionGroupWithDefaults()
	permissionGroupSerializer.SetName(data.Name.ValueString())

	// TODO: implement when available
	// update permissionGroup
	permissionGroupSerializer, httpRes, err := r.client.PermissionsApi.PermissionsGroupsUpdate(context.Background(), state.Name.ValueString()).PermissionGroup(*permissionGroupSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating permissionGroup "+data.Name.String(),
			"Could not update permissionGroup, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// set model based on response
	err = updatePermissionGroupModelFromResponse(permissionGroupSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the permissionGroup model failed", err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *PermissionGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing
	httpRes, err := r.client.PermissionsApi.PermissionsGroupsDestroy(context.Background(), data.Name.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting PermissionGroup",
			"Could not delete PermissionGroup, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

// Create a PermissionGroup based on a terraform plan
func createPermissionGroupFromPlan(plan *PermissionGroupResourceModel, diagnostics *diag.Diagnostics) (*hwmux.PermissionGroup, error) {
	permissionGroupSerializer := hwmux.NewPermissionGroupWithDefaults()
	permissionGroupSerializer.SetName(plan.Name.ValueString())

	return permissionGroupSerializer, nil
}

// Map response body to model and populate Computed attribute values
func updatePermissionGroupModelFromResponse(permissionGroup *hwmux.PermissionGroup, plan *PermissionGroupResourceModel, diagnostics *diag.Diagnostics, client *hwmux.APIClient) (err error) {
	// Map response body to schema and populate Computed attribute values
	plan.Name = types.StringValue(permissionGroup.GetName())
	plan.ID = types.StringValue(strconv.Itoa(int(permissionGroup.GetId())))

	set, diagn := types.SetValueFrom(context.Background(), types.StringType, permissionGroup.GetPermissions())
	if diagn.HasError() {
		diagnostics.Append(diagn...)
		return
	}
	plan.Permissions = set

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	return nil
}
