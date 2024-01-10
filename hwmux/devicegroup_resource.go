package hwmux

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/Silabs-UTF/hwmux-client-golang/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &DeviceGroupResource{}
var _ resource.ResourceWithImportState = &DeviceGroupResource{}

func NewDeviceGroupResource() resource.Resource {
	return &DeviceGroupResource{}
}

// DeviceGroupResource defines the resource implementation.
type DeviceGroupResource struct {
	client *hwmux.APIClient
}

// DeviceGroupResourceModel describes the resource data model.
type DeviceGroupResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Metadata           types.String   `tfsdk:"metadata"`
	Devices            []types.Int64  `tfsdk:"devices"`
	PermissionGroups   []types.String `tfsdk:"permission_groups"`
	Enable_ahs         types.Bool     `tfsdk:"enable_ahs"`
	Enable_ahs_actions types.Bool     `tfsdk:"enable_ahs_actions"`
	LastUpdated        types.String   `tfsdk:"last_updated"`
	Enable_ahs_cas     types.Bool     `tfsdk:"enable_ahs_cas"`
}

func (r *DeviceGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_group"
}

func (r *DeviceGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Device Group resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Device Group identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Device Group name.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`[-a-zA-Z0-9_]+$`), "This field must be a SLUG. No spaces allowed."),
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(100),
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "The metadata of the Device Group.",
				Computed:            true,
				Optional:            true,
			},
			"devices": schema.SetAttribute{
				MarkdownDescription: "The devices that belong to the Device Group.",
				Required:            true,
				ElementType:         types.Int64Type,
			},
			"permission_groups": schema.SetAttribute{
				MarkdownDescription: "Which permission groups can access the resource.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"enable_ahs": schema.BoolAttribute{
				MarkdownDescription: "Enable the Automated Health Service",
				Computed:            true,
				Optional:            true,
			},
			"enable_ahs_actions": schema.BoolAttribute{
				MarkdownDescription: "Allow the Automated Health Service to take DeviceGroups offline when they are unhealthy.",
				Computed:            true,
				Optional:            true,
			},
			"enable_ahs_cas": schema.BoolAttribute{
				MarkdownDescription: "Allow the Automated Health Service to take corrective actions.",
				Computed:            true,
				Optional:            true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the resource.",
				Computed:    true,
			},
		},
	}
}

func (r *DeviceGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeviceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DeviceGroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deviceGroupSerializer, err := createDeviceGroupFromPlan(data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create deviceGroup API request based on plan", err.Error(),
		)
		return
	}

	// create new deviceGroup
	deviceGroupSerializer, httpRes, err := r.client.GroupsAPI.GroupsCreate(context.Background()).DeviceGroupSerializerWithDevicePk(*deviceGroupSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating deviceGroup",
			"Could not create deviceGroup, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	// set model based on response
	err = updateDGModelFromResponse(deviceGroupSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the deviceGroup model failed", err.Error(),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DeviceGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed deviceGroup value from hwmux
	id, _ := strconv.Atoi(data.ID.ValueString())
	deviceGroup, _, err := GetDeviceGroup(r.client, &resp.Diagnostics, int32(id))
	if err != nil {
		return
	}

	// Map response body to model
	data.ID = types.StringValue(strconv.Itoa(int(deviceGroup.GetId())))
	data.Name = types.StringValue(deviceGroup.GetName())
	data.Enable_ahs = types.BoolValue(deviceGroup.GetEnableAhs())
	data.Enable_ahs_actions = types.BoolValue(deviceGroup.GetEnableAhsActions())
	data.Enable_ahs_cas = types.BoolValue(deviceGroup.GetEnableAhsCas())

	err = MarshalMetadataSetError(deviceGroup.GetMetadata(), &resp.Diagnostics, "Device Group", &data.Metadata)
	if err != nil {
		return
	}

	data.Devices = make([]types.Int64, len(deviceGroup.GetDevices()))
	for i, device := range deviceGroup.GetDevices() {
		data.Devices[i] = types.Int64Value(int64(device.GetId()))
	}

	permissionGroups := deviceGroup.GetPermissionGroups()
	data.PermissionGroups = make([]types.String, len(permissionGroups))
	for i, aGroup := range permissionGroups {
		data.PermissionGroups[i] = types.StringValue(aGroup)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DeviceGroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deviceGroupSerializer, err := createDeviceGroupFromPlan(data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create deviceGroup API request based on plan", err.Error(),
		)
		return
	}

	// update deviceGroup
	id, _ := strconv.Atoi(data.ID.ValueString())
	deviceGroupSerializer, httpRes, err := r.client.GroupsAPI.GroupsUpdate(context.Background(), int32(id)).DeviceGroupSerializerWithDevicePk(*deviceGroupSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating deviceGroup "+data.ID.String(),
			"Could not update deviceGroup, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// set model based on response
	err = updateDGModelFromResponse(deviceGroupSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the deviceGroup model failed", err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DeviceGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing
	id, _ := strconv.Atoi(data.ID.ValueString())
	httpRes, err := r.client.GroupsAPI.GroupsDestroy(context.Background(), int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DeviceGroup",
			"Could not delete deviceGroup, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

func (r *DeviceGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func createDeviceGroupFromPlan(plan *DeviceGroupResourceModel, diagnostics *diag.Diagnostics) (*hwmux.DeviceGroupSerializerWithDevicePk, error) {
	deviceGroupSerializer := hwmux.NewDeviceGroupSerializerWithDevicePkWithDefaults()
	deviceGroupSerializer.SetName(plan.Name.ValueString())

	if !plan.Enable_ahs.IsUnknown() {
		deviceGroupSerializer.SetEnableAhs(plan.Enable_ahs.ValueBool())
	}
	if !plan.Enable_ahs_actions.IsUnknown() {
		deviceGroupSerializer.SetEnableAhsActions(plan.Enable_ahs_actions.ValueBool())
	}
	if !plan.Enable_ahs_cas.IsUnknown() {
		deviceGroupSerializer.SetEnableAhsCas(plan.Enable_ahs_cas.ValueBool())
	}

	if !plan.Metadata.IsUnknown() {
		metadata, errorMet := UnmarshalMetadataSetError(plan.Metadata.ValueString(), diagnostics, "deviceGroup")
		if errorMet != nil {
			return nil, errorMet
		}
		deviceGroupSerializer.SetMetadata(*metadata)
	}

	deviceIds := make([]int32, len(plan.Devices))
	for i, device := range plan.Devices {
		deviceIds[i] = int32(device.ValueInt64())
	}

	deviceGroupSerializer.SetDevices(deviceIds)

	permissionList := make([]string, len(plan.PermissionGroups))
	for i, permissionGroup := range plan.PermissionGroups {
		permissionList[i] = permissionGroup.ValueString()
	}

	deviceGroupSerializer.SetPermissionGroups(permissionList)

	return deviceGroupSerializer, nil
}

// Map response body to model and populate Computed attribute values
func updateDGModelFromResponse(deviceGroup *hwmux.DeviceGroupSerializerWithDevicePk, plan *DeviceGroupResourceModel, diagnostics *diag.Diagnostics, client *hwmux.APIClient) (err error) {
	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.Itoa(int(deviceGroup.GetId())))
	plan.Name = types.StringValue(deviceGroup.GetName())
	plan.Enable_ahs = types.BoolValue(deviceGroup.GetEnableAhs())
	plan.Enable_ahs_actions = types.BoolValue(deviceGroup.GetEnableAhsActions())
	plan.Enable_ahs_cas = types.BoolValue(deviceGroup.GetEnableAhsCas())

	err = MarshalMetadataSetError(deviceGroup.GetMetadata(), diagnostics, "deviceGroup", &plan.Metadata)
	if err != nil {
		return
	}

	plan.Devices = make([]types.Int64, len(deviceGroup.GetDevices()))
	for i, device := range deviceGroup.GetDevices() {
		plan.Devices[i] = types.Int64Value(int64(device))
	}

	permissionGroups, err := GetPermissionGroupsForDeviceGroup(client, diagnostics, deviceGroup.GetId())
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
