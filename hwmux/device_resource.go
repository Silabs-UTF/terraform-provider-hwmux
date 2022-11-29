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
	_ resource.Resource                = &deviceResource{}
	_ resource.ResourceWithConfigure   = &deviceResource{}
	_ resource.ResourceWithImportState = &deviceResource{}
)

// NewDeviceResource is a helper function to simplify the provider implementation.
func NewDeviceResource() resource.Resource {
	return &deviceResource{}
}

// deviceResource is the resource implementation.
type deviceResource struct {
	client *hwmux.APIClient
}

type deviceResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Sn_or_name       types.String `tfsdk:"sn_or_name"`
	Is_wstk          types.Bool   `tfsdk:"is_wstk"`
	Uri              types.String `tfsdk:"uri"`
	Online           types.Bool   `tfsdk:"online"`
	Metadata         types.String `tfsdk:"metadata"`
	Part             types.String `tfsdk:"part"`
	Room             types.String `tfsdk:"room"`
	LocationMetadata types.String `tfsdk:"location_metadata"`
	LastUpdated      types.String `tfsdk:"last_updated"`
}

// Configure adds the provider configured client to the resource.
func (r *deviceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*hwmux.APIClient)
}

// Metadata returns the resource type name.
func (r *deviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

// GetSchema defines the schema for the resource.
func (r *deviceResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Manages a device in hwmux.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "The ID of the device.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"sn_or_name": {
				Description: "The name of the device. Must be unique.",
				Type:        types.StringType,
				Optional:    true,
			},
			"is_wstk": {
				Description: "Whether the device is a WSTK.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
			},
			"uri": {
				Description: "The URI or IP address of the device.",
				Type:        types.StringType,
				Optional:    true,
			},
			"online": {
				Description: "Whether the device is online.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
			},
			"metadata": {
				Description: "The metadata of the device.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"part": {
				Description: "The part number of the device.",
				Type:        types.StringType,
				Required:    true,
			},
			"room": {
				Description: "The name of the room the device is in. Must exist in hwmux.",
				Type:        types.StringType,
				Required:    true,
			},
			"last_updated": {
				Description: "Timestamp of the last Terraform update of the device.",
				Type:        types.StringType,
				Computed:    true,
			},
			"location_metadata": {
				Description: "The location metadata of the device.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}, nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan deviceResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	writeOnlyDevice, err := createDeviceFromPlan(&plan, &resp.Diagnostics)
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
			"Error creating device",
			"Could not create device, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	// set model based on response
	err = updateModelFromResponse(writeOnlyDevice, &plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the device model failed", err.Error(),
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
func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed device value from hwmux
	id, _ := strconv.Atoi(state.ID.ValueString())
	device, _, err := GetDevice(r.client, &resp.Diagnostics, int32(id))
	if err != nil {
		return
	}

	// Map response body to model
	state.ID = types.StringValue(strconv.Itoa(int(device.GetId())))
	state.Sn_or_name = types.StringValue(device.GetSnOrName())
	state.Is_wstk = types.BoolValue(device.GetIsWstk())
	state.Uri = types.StringValue(device.GetUri())
	state.Online = types.BoolValue(device.GetOnline())

	location, _, err := GetDeviceLocation(r.client, &resp.Diagnostics, device.GetId())
	if err == nil {
		state.Room = types.StringValue(location.Room.GetName())
	}

	err = MarshalMetadataSetError(location.GetMetadata(), &resp.Diagnostics, "location", &state.LocationMetadata)
	if err != nil {
		return
	}

	err = MarshalMetadataSetError(device.GetMetadata(), &resp.Diagnostics, "device", &state.Metadata)
	if err != nil {
		return
	}

	state.Part = types.StringValue(device.Part.GetPartNo())

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deviceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	writeOnlyDevice, err := createDeviceFromPlan(&plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create device API request based on plan", err.Error(),
		)
		return
	}

	// update device
	id, _ := strconv.Atoi(plan.ID.ValueString())
	writeOnlyDevice, httpRes, err := r.client.DevicesApi.DevicesUpdate(context.Background(), int32(id)).WriteOnlyDevice(*writeOnlyDevice).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating device "+plan.ID.String(),
			"Could not update device, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// set model based on response
	err = updateModelFromResponse(writeOnlyDevice, &plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the device model failed", err.Error(),
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
func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	id, _ := strconv.Atoi(state.ID.ValueString())
	httpRes, err := r.client.DevicesApi.DevicesDestroy(context.Background(), int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Device",
			"Could not delete device, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

func (r *deviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a writeOnlyDevice based on a terraform plan
func createDeviceFromPlan(plan *deviceResourceModel, diagnostics *diag.Diagnostics) (*hwmux.WriteOnlyDevice, error) {
	writeOnlyDevice := hwmux.NewWriteOnlyDeviceWithDefaults()
	writeOnlyDevice.SetPart(plan.Part.ValueString())

	if !plan.Online.IsUnknown() {
		writeOnlyDevice.SetOnline(plan.Online.ValueBool())
	} else {
		writeOnlyDevice.SetOnline(true)
	}
	if !plan.Is_wstk.IsUnknown() {
		writeOnlyDevice.SetIsWstk(plan.Is_wstk.ValueBool())
	}
	if !plan.Sn_or_name.IsNull() {
		sn_or_name := plan.Sn_or_name.ValueString()
		writeOnlyDevice.SnOrName.Set(&sn_or_name)
	}
	if !plan.Uri.IsNull() {
		writeOnlyDevice.SetUri(plan.Uri.ValueString())
	}
	if !plan.Metadata.IsUnknown() {
		metadata, errorMet := UnmarshalMetadataSetError(plan.Metadata.ValueString(), diagnostics, "device")
		if errorMet != nil {
			return nil, errorMet
		}
		writeOnlyDevice.SetMetadata(*metadata)
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

	return writeOnlyDevice, nil
}

// Map response body to model and populate Computed attribute values
func updateModelFromResponse(device *hwmux.WriteOnlyDevice, plan *deviceResourceModel, diagnostics *diag.Diagnostics) (err error) {
	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.Itoa(int(device.GetId())))
	plan.Sn_or_name = types.StringValue(device.GetSnOrName())
	plan.Is_wstk = types.BoolValue(device.GetIsWstk())
	plan.Uri = types.StringValue(device.GetUri())
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
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	return nil
}
