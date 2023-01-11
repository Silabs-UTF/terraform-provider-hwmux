package hwmux

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Silabs-UTF/hwmux-client-golang"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
	client *hwmux.APIClient
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Username         types.String   `tfsdk:"username"`
	FirstName        types.String   `tfsdk:"first_name"`
	LastName         types.String   `tfsdk:"last_name"`
	Email            types.String   `tfsdk:"email"`
	IsStaff          types.Bool     `tfsdk:"is_staff"`
	IsSuperuser      types.Bool     `tfsdk:"is_superuser"`
	PermissionGroups []types.String `tfsdk:"permission_groups"`
	LastUpdated      types.String   `tfsdk:"last_updated"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Username.",
			},
			"first_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "User first name.",
			},
			"last_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "User last name.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "User last name.",
			},
			"is_staff": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the user is a staff user.",
			},
			"is_superuser": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the user is a super user.",
			},
			"permission_groups": schema.SetAttribute{
				MarkdownDescription: "Which permission groups can access the resource.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the resource.",
				Computed:    true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userSerializer, err := createUserFromPlan(data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create user API request based on plan", err.Error(),
		)
		return
	}

	// create new user
	userSerializer, httpRes, err := r.client.UserApi.UserCreate(context.Background()).LoggedInUser(*userSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			"Could not create user, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// process group membership
	err = processUserPermissions(userSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}

	// Map response body to schema and populate Computed attribute values
	// set model based on response
	err = updateUserModelFromResponse(userSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the user model failed", err.Error(),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed user value from hwmux
	user, _, err := GetUser(r.client, &resp.Diagnostics, data.ID.ValueString())
	if err != nil {
		return
	}

	err = updateUserModelFromResponse(user, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the user model failed", err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userSerializer, err := createUserFromPlan(data, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create user API request based on plan", err.Error(),
		)
		return
	}

	// update user
	userSerializer, httpRes, err := r.client.UserApi.UserUpdate(context.Background(), data.ID.ValueString()).LoggedInUser(*userSerializer).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating user "+data.ID.String(),
			"Could not update user, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// process group membership
	err = processUserPermissions(userSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		return
	}

	// set model based on response
	err = updateUserModelFromResponse(userSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the user model failed", err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing
	httpRes, err := r.client.UserApi.UserDestroy(context.Background(), data.ID.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting User",
			"Could not delete user, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Create a User based on a terraform plan
func createUserFromPlan(plan *UserResourceModel, diagnostics *diag.Diagnostics) (*hwmux.LoggedInUser, error) {
	userSerializer := hwmux.NewLoggedInUserWithDefaults()
	userSerializer.SetUsername(plan.Username.ValueString())
	userSerializer.SetFirstName(plan.FirstName.ValueString())
	userSerializer.SetLastName(plan.LastName.ValueString())
	userSerializer.SetEmail(plan.Email.ValueString())

	return userSerializer, nil
}

// Map response body to model and populate Computed attribute values
func updateUserModelFromResponse(user *hwmux.LoggedInUser, plan *UserResourceModel, diagnostics *diag.Diagnostics, client *hwmux.APIClient) (err error) {
	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.Itoa(int(user.GetId())))
	plan.Username = types.StringValue(user.GetUsername())
	plan.FirstName = types.StringValue(user.GetFirstName())
	plan.LastName = types.StringValue(user.GetLastName())
	plan.Email = types.StringValue(user.GetEmail())
	plan.IsStaff = types.BoolValue(user.GetIsStaff())
	plan.IsSuperuser = types.BoolValue(user.GetIsSuperuser())

	plan.PermissionGroups = make([]types.String, len(user.GetGroups()))
	for i, aGroup := range user.GetGroups() {
		plan.PermissionGroups[i] = types.StringValue(aGroup)
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	return nil
}
