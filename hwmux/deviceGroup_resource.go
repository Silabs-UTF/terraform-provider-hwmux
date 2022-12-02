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
	_ resource.Resource                = &deviceGroupResource{}
	_ resource.ResourceWithConfigure   = &deviceGroupResource{}
	_ resource.ResourceWithImportState = &deviceGroupResource{}
)

// NewDeviceGroupResource is a helper function to simplify the provider implementation.
func NewDeviceGroupResource() resource.Resource {
	return &deviceGroupResource{}
}

// deviceGroupResource is the resource implementation.
type deviceGroupResource struct {
	client *hwmux.APIClient
}

type deviceGroupResourceModel struct {
	ID               types.String                 `tfsdk:"id"`
	Name             types.String                 `tfsdk:"name"`
	Metadata         types.String                 `tfsdk:"metadata"`
	Devices          []types.Int64   `tfsdk:"devices"`
	PermissionGroups []types.String `tfsdk:"permission_groups"`
	LastUpdated      types.String                 `tfsdk:"last_updated"`
}

// Configure adds the provider configured client to the resource.
func (r *deviceGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*hwmux.APIClient)
}

// Metadata returns the resource type name.
func (r *deviceGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deviceGroup"
}

// GetSchema defines the schema for the resource.
func (r *deviceGroupResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Manages a deviceGroup in hwmux.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "The ID of the deviceGroup.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				Description: "The name of the deviceGroup. Must be unique.",
				Type:        types.StringType,
				Required:    true,
			},
			"metadata": {
				Description: "The metadata of the deviceGroup.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"devices": {
				Description: "The IDs of the devices that belong to the deviceGroup.",
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
func (r *deviceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan deviceGroupResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceGroupSerializer, err := createDeviceGroupFromPlan(&plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create deviceGroup API request based on plan", err.Error(),
		)
		return
	}

	// create new deviceGroup
	deviceGroupSerializer, httpRes, err := r.client.GroupsApi.GroupsCreate(context.Background()).DeviceGroupSerializerWithDevicePk(*deviceGroupSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deviceGroup",
			"Could not create deviceGroup, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	// set model based on response
	err = updateDGModelFromResponse(deviceGroupSerializer, &plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the deviceGroup model failed", err.Error(),
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
func (r *deviceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed deviceGroup value from hwmux
	id, _ := strconv.Atoi(state.ID.ValueString())
	deviceGroup, _, err := GetDeviceGroup(r.client, &resp.Diagnostics, int32(id))
	if err != nil {
		return
	}

	// Map response body to model
	state.ID = types.StringValue(strconv.Itoa(int(deviceGroup.GetId())))
	state.Name = types.StringValue(deviceGroup.GetName())

	err = MarshalMetadataSetError(deviceGroup.GetMetadata(), &resp.Diagnostics, "deviceGroup", &state.Metadata)
	if err != nil {
		return
	}

	state.Devices = []types.Int64{}
	for _, device := range deviceGroup.GetDevices() {
		state.Devices = append(state.Devices, types.Int64Value(int64(device.GetId())))
	}

	// TODO: implement permission groups Read once API is available

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *deviceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deviceGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceGroupSerializer, err := createDeviceGroupFromPlan(&plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create deviceGroup API request based on plan", err.Error(),
		)
		return
	}

	// update deviceGroup
	id, _ := strconv.Atoi(plan.ID.ValueString())
	deviceGroupSerializer, httpRes, err := r.client.GroupsApi.GroupsUpdate(context.Background(), int32(id)).DeviceGroupSerializerWithDevicePk(*deviceGroupSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating deviceGroup "+plan.ID.String(),
			"Could not update deviceGroup, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// set model based on response
	err = updateDGModelFromResponse(deviceGroupSerializer, &plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the deviceGroup model failed", err.Error(),
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
func (r *deviceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	id, _ := strconv.Atoi(state.ID.ValueString())
	httpRes, err := r.client.GroupsApi.GroupsDestroy(context.Background(), int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DeviceGroup",
			"Could not delete deviceGroup, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

func (r *deviceGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a writeOnlyDeviceGroup based on a terraform plan
func createDeviceGroupFromPlan(plan *deviceGroupResourceModel, diagnostics *diag.Diagnostics) (*hwmux.DeviceGroupSerializerWithDevicePk, error) {
	deviceGroupSerializer := hwmux.NewDeviceGroupSerializerWithDevicePkWithDefaults()
	deviceGroupSerializer.SetName(plan.Name.ValueString())
	
	if !plan.Metadata.IsUnknown() {
		metadata, errorMet := UnmarshalMetadataSetError(plan.Metadata.ValueString(), diagnostics, "deviceGroup")
		if errorMet != nil {
			return nil, errorMet
		}
		deviceGroupSerializer.SetMetadata(*metadata)
	}

	deviceIds := []int32{}
	for _, device := range plan.Devices {
		deviceIds = append(deviceIds, int32(device.ValueInt64()))
	}

	deviceGroupSerializer.SetDevices(deviceIds)
	
	permissionList := []string{}
	for _, permissionGroup := range plan.PermissionGroups {
		permissionList = append(permissionList, permissionGroup.ValueString())
	}

	deviceGroupSerializer.SetPermissionGroups(permissionList)

	return deviceGroupSerializer, nil
}

// Map response body to model and populate Computed attribute values
func updateDGModelFromResponse(deviceGroup *hwmux.DeviceGroupSerializerWithDevicePk, plan *deviceGroupResourceModel, diagnostics *diag.Diagnostics) (err error) {
	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.Itoa(int(deviceGroup.GetId())))
	plan.Name = types.StringValue(deviceGroup.GetName())

	err = MarshalMetadataSetError(deviceGroup.GetMetadata(), diagnostics, "deviceGroup", &plan.Metadata)
	if err != nil {
		return
	}

	plan.Devices = []types.Int64{}
	for _, device := range deviceGroup.GetDevices() {
		plan.Devices = append(plan.Devices, types.Int64Value(int64(device)))
	}

	// TODO: Implement device group read when available

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	return nil
}
