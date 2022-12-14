package hwmux

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"stash.silabs.com/iot_infra_sw/hwmux-client-golang"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &LabelResource{}
var _ resource.ResourceWithImportState = &LabelResource{}

func NewLabelResource() resource.Resource {
	return &LabelResource{}
}

// LabelResource defines the resource implementation.
type LabelResource struct {
	client *hwmux.APIClient
}

// LabelResourceModel describes the resource data model.
type LabelResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	Metadata         types.String   `tfsdk:"metadata"`
	DeviceGroups     []types.Int64  `tfsdk:"device_groups"`
	PermissionGroups []types.String `tfsdk:"permission_groups"`
	LastUpdated      types.String   `tfsdk:"last_updated"`
}

func (r *LabelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (r *LabelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Label resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Label identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Label name.",
			},
			"metadata": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Label metadata.",
			},
			"permission_groups": schema.SetAttribute{
				MarkdownDescription: "Which permission groups can access the resource.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"device_groups": schema.SetAttribute{
				MarkdownDescription: "The IDs of the deviceGroups that belong to the label.",
				ElementType:         types.Int64Type,
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the resource.",
				Computed:    true,
			},
		},
	}
}

func (r *LabelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LabelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *LabelResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	labelSerializer, err := createLabelFromPlan(data, &resp.Diagnostics)
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
	err = updateLabelModelFromResponse(labelSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the label model failed", err.Error(),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LabelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *LabelResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed label value from hwmux
	id, _ := strconv.Atoi(data.ID.ValueString())
	label, _, err := GetLabel(r.client, &resp.Diagnostics, int32(id))
	if err != nil {
		return
	}

	// Map response body to model
	data.ID = types.StringValue(strconv.Itoa(int(label.GetId())))
	data.Name = types.StringValue(label.GetName())

	err = MarshalMetadataSetError(label.GetMetadata(), &resp.Diagnostics, "label", &data.Metadata)
	if err != nil {
		return
	}

	data.DeviceGroups = make([]types.Int64, len(label.GetDeviceGroups()))
	for i, deviceGroup := range label.GetDeviceGroups() {
		data.DeviceGroups[i] = types.Int64Value(int64(deviceGroup))
	}

	permissionGroups, err := GetPermissionGroupsForLabel(r.client, &resp.Diagnostics, label.GetId())
	if err != nil {
		return
	}
	data.PermissionGroups = make([]types.String, len(permissionGroups))
	for i, aGroup := range permissionGroups {
		data.PermissionGroups[i] = types.StringValue(aGroup)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LabelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *LabelResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	labelSerializer, err := createLabelFromPlan(data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create label API request based on plan", err.Error(),
		)
		return
	}

	// update label
	id, _ := strconv.Atoi(data.ID.ValueString())
	labelSerializer, httpRes, err := r.client.LabelsApi.LabelsUpdate(context.Background(), int32(id)).LabelSerializerWithPermissions(*labelSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating label "+data.ID.String(),
			"Could not update label, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// set model based on response
	err = updateLabelModelFromResponse(labelSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the label model failed", err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *LabelResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing
	id, _ := strconv.Atoi(data.ID.ValueString())
	httpRes, err := r.client.LabelsApi.LabelsDestroy(context.Background(), int32(id)).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Label",
			"Could not delete label, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

func (r *LabelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a Label based on a terraform plan
func createLabelFromPlan(plan *LabelResourceModel, diagnostics *diag.Diagnostics) (*hwmux.LabelSerializerWithPermissions, error) {
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
func updateLabelModelFromResponse(label *hwmux.LabelSerializerWithPermissions, plan *LabelResourceModel, diagnostics *diag.Diagnostics, client *hwmux.APIClient) (err error) {
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
