package hwmux

import (
	"context"
	"fmt"
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
var _ resource.Resource = &DeviceResource{}
var _ resource.ResourceWithImportState = &DeviceResource{}

func NewDeviceResource() resource.Resource {
	return &DeviceResource{}
}

// DeviceResource defines the resource implementation.
type DeviceResource struct {
	client *hwmux.APIClient
}

// DeviceResourceModel describes the resource data model.
type DeviceResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Sn_or_name       types.String   `tfsdk:"sn_or_name"`
	Is_wstk          types.Bool     `tfsdk:"is_wstk"`
	Uri              types.String   `tfsdk:"uri"`
	Online           types.Bool     `tfsdk:"online"`
	Metadata         types.String   `tfsdk:"metadata"`
	Part             types.String   `tfsdk:"part"`
	Wstk_part        types.String   `tfsdk:"wstk_part"`
	Room             types.String   `tfsdk:"room"`
	LocationMetadata types.String   `tfsdk:"location_metadata"`
	PermissionGroups []types.String `tfsdk:"permission_groups"`
	LastUpdated      types.String   `tfsdk:"last_updated"`
	Source           types.String   `tfsdk:"source"`
	Socketed_chip    types.String   `tfsdk:"socketed_chip"`
}

func (r *DeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *DeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Device resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Device identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sn_or_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Device name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(255),
				},
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "The URI or IP address of the device.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(255),
				},
			},
			"part": schema.StringAttribute{
				MarkdownDescription: "The part number of the device.",
				Required:            true,
			},
			"room": schema.StringAttribute{
				MarkdownDescription: "The room where the device is.",
				Required:            true,
			},
			"is_wstk": schema.BoolAttribute{
				MarkdownDescription: "If the device is a WSTK.",
				Computed:            true,
				Optional:            true,
			},
			"wstk_part": schema.StringAttribute{
				MarkdownDescription: "The part number of the WSTK the device is on.",
				Optional:            true,
			},
			"online": schema.BoolAttribute{
				MarkdownDescription: "If the device is online.",
				Computed:            true,
				Optional:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "The metadata of the device.",
				Computed:            true,
				Optional:            true,
			},
			"location_metadata": schema.StringAttribute{
				MarkdownDescription: "The location metadata of the device.",
				Computed:            true,
				Optional:            true,
			},
			"permission_groups": schema.SetAttribute{
				MarkdownDescription: "Which permission groups can access the resource.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the resource.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "The source where the device was created.",
				Computed:    true,
			},
			"socketed_chip": schema.StringAttribute{
				MarkdownDescription: "The socket chip detail of the device.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(255),
				},
			},
		},
	}
}

func (r *DeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DeviceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	writeOnlyDevice, err := createDeviceFromPlan(data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create device API request based on plan", err.Error(),
		)
		return
	}

	// create new device
	writeOnlyDevice, httpRes, err := r.client.DevicesApi.DevicesCreate(context.Background()).WriteOnlyDevice(*writeOnlyDevice).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating device %s", data.Sn_or_name.String()),
			fmt.Sprintf("Could not create device %s, unexpected error. Error %s \n Response body %s", data.Sn_or_name.String(), err.Error(), BodyToString(&httpRes.Body)),
		)
		return
	}

	// Handle the online field, which is remapped to status
	if !data.Online.IsUnknown() && !data.Online.ValueBool() {
		statusReq, err := r.setDeviceStatusFromPlan(&resp.Diagnostics, writeOnlyDevice.GetId(), hwmux.DISABLED)
		if err != nil {
			resp.Diagnostics.AddError("Error updating device status", err.Error())
			return
		}

		writeOnlyDevice.SetOnline(data.Online.ValueBool())
		writeOnlyDevice.SetStatus(statusReq.Status)
	}

	// Map response body to schema and populate Computed attribute values
	// set model based on response
	err = updateDeviceModelFromResponse(writeOnlyDevice, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Updating the device model failed %s", data.Sn_or_name.String()), err.Error(),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DeviceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed device value from hwmux
	id, _ := strconv.Atoi(data.ID.ValueString())
	device, _, err := GetDevice(r.client, &resp.Diagnostics, int32(id))
	if err != nil {
		// add diagnostic error with the expected ID
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading device %d", id),
			fmt.Sprintf("Could not read device %d, unexpected error. Error %s", id, err.Error()),
		)
		return
	}

	// Map response body to model
	data.ID = types.StringValue(strconv.Itoa(int(device.GetId())))
	if device.GetSnOrName() != "" {
		data.Sn_or_name = types.StringValue(device.GetSnOrName())
	} else {
		data.Sn_or_name = types.StringNull()
	}
	data.Source = types.StringValue(string(device.GetSource()))
	if device.GetSource() != "" {
		data.Source = types.StringValue(string(device.GetSource()))
	} else {
		data.Source = types.StringNull()
	}
	data.Is_wstk = types.BoolValue(device.GetIsWstk())
	if device.GetWstkPart() != "" {
		data.Wstk_part = types.StringValue(device.GetWstkPart())
	} else {
		data.Wstk_part = types.StringNull()
	}
	if device.GetUri() != "" {
		data.Uri = types.StringValue(device.GetUri())
	} else {
		data.Uri = types.StringNull()
	}
	if device.GetSocketedChip() != "" {
		data.Socketed_chip = types.StringValue(string(device.GetSocketedChip()))
	} else {
		data.Socketed_chip = types.StringNull()
	}
	data.Online = types.BoolValue(device.GetOnline())

	location, _, err := GetDeviceLocation(r.client, &resp.Diagnostics, device.GetId())
	if err == nil {
		data.Room = types.StringValue(location.Room.GetName())
	}

	err = MarshalMetadataSetError(location.GetMetadata(), &resp.Diagnostics, "location", &data.LocationMetadata)
	if err != nil {
		return
	}

	err = MarshalMetadataSetError(device.GetMetadata(), &resp.Diagnostics, "device", &data.Metadata)
	if err != nil {
		return
	}

	data.Part = types.StringValue(device.Part.GetPartNo())

	permissionGroups := device.GetPermissionGroups()
	data.PermissionGroups = make([]types.String, len(permissionGroups))
	for i, aGroup := range permissionGroups {
		data.PermissionGroups[i] = types.StringValue(aGroup)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DeviceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	writeOnlyDevice, err := createDeviceFromPlan(data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create device API request based on plan", err.Error(),
		)
		return
	}

	if data.Source.ValueString() != "TERRAFORM" {
		writeOnlyDevice.SetSource(hwmux.SOURCEENUM_TERRAFORM)
	}

	// update device
	id, _ := strconv.Atoi(data.ID.ValueString())
	writeOnlyDevice, httpRes, err := r.client.DevicesApi.DevicesUpdate(context.Background(), int32(id)).WriteOnlyDevice(*writeOnlyDevice).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating device %d", id), fmt.Sprintf("Could not update device %d, unexpected error. Error %s \n Response body %s", id, err.Error(), BodyToString(&httpRes.Body)),
		)
		return
	}

	// Handle the online field, which is remapped to status
	//  will only make a change if there is a difference between the API-provided value and the desired one
	if (data.Online.IsUnknown() && !writeOnlyDevice.GetOnline()) || (!data.Online.IsUnknown() && data.Online.ValueBool() != writeOnlyDevice.GetOnline()) {
		status := hwmux.ACTIVE
		if !data.Online.IsUnknown() && !data.Online.ValueBool() {
			status = hwmux.DISABLED
		}

		statusReq, err := r.setDeviceStatusFromPlan(&resp.Diagnostics, writeOnlyDevice.GetId(), status)
		if err != nil {
			resp.Diagnostics.AddError("Error updating device status", err.Error())
			return
		}

		writeOnlyDevice.SetOnline(data.Online.ValueBool())
		writeOnlyDevice.SetStatus(statusReq.Status)
	}

	// set model based on response
	err = updateDeviceModelFromResponse(writeOnlyDevice, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the device model failed", err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DeviceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// if it's offline, set it back online to remove the reservation for the offline status
	if !data.Online.ValueBool() {
		id, _ := strconv.Atoi(data.ID.ValueString())
		r.setDeviceStatusFromPlan(&resp.Diagnostics, int32(id), hwmux.ACTIVE)
	}

	// Delete existing
	id, _ := strconv.Atoi(data.ID.ValueString())
	httpRes, err := r.client.DevicesApi.DevicesDestroy(context.Background(), int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Device",
			"Could not delete device, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

func (r *DeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DeviceResource) setDeviceStatusFromPlan(diagnostics *diag.Diagnostics, id int32, status hwmux.StatusEnum) (*hwmux.ResourceStatusRequest, error) {

	statusRequest := hwmux.NewResourceStatusRequestWithDefaults()
	statusRequest.SetComment("Disabled via Terraform")
	statusRequest.SetStatus(status)

	resourceStatRequest, httpRes, err := r.client.DevicesApi.DevicesStatusCreate(context.Background(), id).ResourceStatusRequest(*statusRequest).Execute()
	if err != nil {
		diagnostics.AddError(
			"Error setting device status "+strconv.Itoa(int(id)),
			"Could not update device, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return resourceStatRequest, err
	}
	return resourceStatRequest, nil
}

// Create a writeOnlyDevice based on a terraform plan
func createDeviceFromPlan(plan *DeviceResourceModel, diagnostics *diag.Diagnostics) (*hwmux.WriteOnlyDevice, error) {
	writeOnlyDevice := hwmux.NewWriteOnlyDeviceWithDefaults()

	writeOnlyDevice.SetPart(plan.Part.ValueString())
	writeOnlyDevice.SetSource(hwmux.SOURCEENUM_TERRAFORM)

	if !plan.Wstk_part.IsUnknown() {
		writeOnlyDevice.SetWstkPart(plan.Wstk_part.ValueString())
	}

	if !plan.Is_wstk.IsUnknown() {
		writeOnlyDevice.SetIsWstk(plan.Is_wstk.ValueBool())
	}
	if plan.Sn_or_name.IsUnknown() {
		writeOnlyDevice.SetSnOrNameNil()
	} else {
		sn_or_name := plan.Sn_or_name.ValueString()
		writeOnlyDevice.SnOrName.Set(&sn_or_name)
	}
	if plan.Uri.IsUnknown() {
		writeOnlyDevice.SetUriNil()
	} else {
		writeOnlyDevice.SetUri(plan.Uri.ValueString())
	}
	if !plan.Metadata.IsUnknown() {
		metadata, errorMet := UnmarshalMetadataSetError(plan.Metadata.ValueString(), diagnostics, "device")
		if errorMet != nil {
			return nil, errorMet
		}
		writeOnlyDevice.SetMetadata(*metadata)
	}
	if plan.Socketed_chip.IsUnknown() {
		writeOnlyDevice.SetSocketedChip("")
	} else {
		socketed_chip := plan.Socketed_chip.ValueString()
		writeOnlyDevice.SetSocketedChip(socketed_chip)
	}

	location := hwmux.NewLocationSerializerWriteOnlyWithDefaults()
	location.SetRoom(plan.Room.ValueString())
	if !plan.LocationMetadata.IsUnknown() {
		metadata, errorMet := UnmarshalMetadataSetError(plan.LocationMetadata.ValueString(), diagnostics, "location")
		if errorMet != nil {
			return nil, errorMet
		}
		location.SetMetadata(*metadata)
	}

	writeOnlyDevice.SetLocation(*location)

	permissionList := make([]string, len(plan.PermissionGroups))
	for i, permissionGroup := range plan.PermissionGroups {
		permissionList[i] = permissionGroup.ValueString()
	}

	writeOnlyDevice.SetPermissionGroups(permissionList)

	return writeOnlyDevice, nil
}

// Map response body to model and populate Computed attribute values
func updateDeviceModelFromResponse(device *hwmux.WriteOnlyDevice, plan *DeviceResourceModel, diagnostics *diag.Diagnostics,
	client *hwmux.APIClient) (err error) {
	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.Itoa(int(device.GetId())))
	if device.GetSnOrName() != "" {
		plan.Sn_or_name = types.StringValue(device.GetSnOrName())
	} else {
		plan.Sn_or_name = types.StringNull()
	}
	plan.Source = types.StringValue(string(device.GetSource()))
	if device.GetSource() != "" {
		plan.Source = types.StringValue(string(device.GetSource()))
	} else {
		plan.Source = types.StringNull()
	}
	plan.Is_wstk = types.BoolValue(device.GetIsWstk())
	if device.GetUri() != "" {
		plan.Uri = types.StringValue(device.GetUri())
	} else {
		plan.Uri = types.StringNull()
	}
	plan.Online = types.BoolValue(device.GetOnline())
	plan.Room = types.StringValue(plan.Room.ValueString())

	err = MarshalMetadataSetError(device.GetMetadata(), diagnostics, "device", &plan.Metadata)
	if err != nil {
		return
	}

	err = MarshalMetadataSetError(device.Location.GetMetadata(), diagnostics, "location", &plan.LocationMetadata)
	if err != nil {
		return
	}

	plan.Part = types.StringValue(device.Part)

	if device.GetWstkPart() != "" {
		plan.Wstk_part = types.StringValue(device.GetWstkPart())
	} else {
		plan.Wstk_part = types.StringNull()
	}

	if device.GetSocketedChip() != "" {
		plan.Socketed_chip = types.StringValue(string(device.GetSocketedChip()))
	} else {
		plan.Socketed_chip = types.StringNull()
	}

	permissionGroups, err := GetPermissionGroupsForDevice(client, diagnostics, device.GetId())
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
