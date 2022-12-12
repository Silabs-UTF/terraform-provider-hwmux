package hwmux

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"stash.silabs.com/iot_infra_sw/hwmux-client-golang"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &labelResource{}
	_ resource.ResourceWithConfigure   = &labelResource{}
	_ resource.ResourceWithImportState = &labelResource{}
)

// NewLabelResource is a helper function to simplify the provider implementation.
func NewLabelResource() resource.Resource {
	return &labelResource{}
}

// labelResource is the resource implementation.
type labelResource struct {
	client *hwmux.APIClient
}

type labelResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	Metadata         types.String   `tfsdk:"metadata"`
	DeviceGroups     []types.Int64  `tfsdk:"device_groups"`
	PermissionGroups []types.String `tfsdk:"permission_groups"`
	LastUpdated      types.String   `tfsdk:"last_updated"`
}

// Configure adds the provider configured client to the resource.
func (r *labelResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*hwmux.APIClient)
}

// Metadata returns the resource type name.
func (r *labelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

// GetSchema defines the schema for the resource.
func (r *labelResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Manages a label in hwmux.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "The ID of the label.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				Description: "The name of the label. Must be unique.",
				Type:        types.StringType,
				Required:    true,
			},
			"metadata": {
				Description: "The metadata of the label.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"device_groups": {
				Description: "The IDs of the deviceGroups that belong to the label.",
				Required:    true,
				Type: types.SetType{
					ElemType: types.Int64Type,
				},
			},
			"permission_groups": {
				Description: "Which permission groups can access the resource.",
				Required:    true,
				Type: types.SetType{
					ElemType: types.StringType,
				},
			},
			"last_updated": {
				Description: "Timestamp of the last Terraform update of the device.",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}, nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *labelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan labelResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	labelSerializer, err := createLabelFromPlan(&plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create label API request based on plan", err.Error(),
		)
		return
	}

	// create new label
	labelSerializer, httpRes, err := r.client.LabelsApi.LabelsCreate(context.Background()).LabelSerializerWithPermissions(*labelSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating label",
			"Could not create label, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	// set model based on response
	err = updateLabelModelFromResponse(labelSerializer, &plan, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the label model failed", err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *labelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state labelResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed label value from hwmux
	id, _ := strconv.Atoi(state.ID.ValueString())
	label, _, err := GetLabel(r.client, &resp.Diagnostics, int32(id))
	if err != nil {
		return
	}

	// Map response body to model
	state.ID = types.StringValue(strconv.Itoa(int(label.GetId())))
	state.Name = types.StringValue(label.GetName())

	err = MarshalMetadataSetError(label.GetMetadata(), &resp.Diagnostics, "label", &state.Metadata)
	if err != nil {
		return
	}

	state.DeviceGroups = make([]types.Int64, len(label.GetDeviceGroups()))
	for i, deviceGroup := range label.GetDeviceGroups() {
		state.DeviceGroups[i] = types.Int64Value(int64(deviceGroup))
	}

	permissionGroups, err := GetPermissionGroupsForLabel(r.client, &resp.Diagnostics, label.GetId())
	if err != nil {
		return
	}
	state.PermissionGroups = make([]types.String, len(permissionGroups))
	for i, aGroup := range permissionGroups {
		state.PermissionGroups[i] = types.StringValue(aGroup)
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *labelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan labelResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	labelSerializer, err := createLabelFromPlan(&plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create label API request based on plan", err.Error(),
		)
		return
	}

	// update label
	id, _ := strconv.Atoi(plan.ID.ValueString())
	labelSerializer, httpRes, err := r.client.LabelsApi.LabelsUpdate(context.Background(), int32(id)).LabelSerializerWithPermissions(*labelSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating label "+plan.ID.String(),
			"Could not update label, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// set model based on response
	err = updateLabelModelFromResponse(labelSerializer, &plan, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the label model failed", err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *labelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state labelResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	id, _ := strconv.Atoi(state.ID.ValueString())
	httpRes, err := r.client.LabelsApi.LabelsDestroy(context.Background(), int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Label",
			"Could not delete label, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

func (r *labelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a writeOnlyLabel based on a terraform plan
func createLabelFromPlan(plan *labelResourceModel, diagnostics *diag.Diagnostics) (*hwmux.LabelSerializerWithPermissions, error) {
	labelSerializer := hwmux.NewLabelSerializerWithPermissionsWithDefaults()
	labelSerializer.SetName(plan.Name.ValueString())

	if !plan.Metadata.IsUnknown() {
		metadata, errorMet := UnmarshalMetadataSetError(plan.Metadata.ValueString(), diagnostics, "label")
		if errorMet != nil {
			return nil, errorMet
		}
		labelSerializer.SetMetadata(*metadata)
	}

	deviceGroupIds := make([]int32, len(plan.DeviceGroups))
	for i, device := range plan.DeviceGroups {
		deviceGroupIds[i] = int32(device.ValueInt64())
	}

	labelSerializer.SetDeviceGroups(deviceGroupIds)

	permissionList := make([]string, len(plan.PermissionGroups))
	for i, permissionGroup := range plan.PermissionGroups {
		permissionList[i] = permissionGroup.ValueString()
	}

	labelSerializer.SetPermissionGroups(permissionList)

	return labelSerializer, nil
}

// Map response body to model and populate Computed attribute values
func updateLabelModelFromResponse(label *hwmux.LabelSerializerWithPermissions, plan *labelResourceModel, diagnostics *diag.Diagnostics, client *hwmux.APIClient) (err error) {
	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.Itoa(int(label.GetId())))
	plan.Name = types.StringValue(label.GetName())

	err = MarshalMetadataSetError(label.GetMetadata(), diagnostics, "label", &plan.Metadata)
	if err != nil {
		return
	}

	plan.DeviceGroups = make([]types.Int64, len(label.GetDeviceGroups()))
	for i, deviceGroup := range label.GetDeviceGroups() {
		plan.DeviceGroups[i] = types.Int64Value(int64(deviceGroup))
	}

	permissionGroups, err := GetPermissionGroupsForLabel(client, diagnostics, label.GetId())
	if err != nil {
		return
	}
	plan.PermissionGroups = make([]types.String, len(permissionGroups))
	for i, aGroup := range permissionGroups {
		plan.PermissionGroups[i] = types.StringValue(aGroup)
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	return nil
}
