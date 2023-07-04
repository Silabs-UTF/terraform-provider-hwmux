package hwmux

import (
	"context"
	"fmt"
	"time"

	"github.com/Silabs-UTF/hwmux-client-golang/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &TokenResource{}

func NewTokenResource() resource.Resource {
	return &TokenResource{}
}

// TokenResource defines the resource implementation.
type TokenResource struct {
	client *hwmux.APIClient
}

// TokenResourceModel describes the resource data model.
type TokenResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Token       types.String `tfsdk:"token"`
	UserId      types.String `tfsdk:"user_id"`
	DateCreated types.String `tfsdk:"date_created"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *TokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

func (r *TokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Token resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Token identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The user Id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The token.",
			},
			"date_created": schema.StringAttribute{
				Description: "Timestamp of the time the token was created.",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the resource.",
				Computed:    true,
			},
		},
	}
}

func (r *TokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// create new token
	tokenSerializer, httpRes, err := r.client.UserApi.UserTokenCreate(context.Background(), data.UserId.ValueString()).Execute()
	data.ID = types.StringValue(tokenSerializer.GetKey())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating token",
			"Could not create token, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	// set model based on response
	err = updateTokenModelFromResponse(tokenSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the token model failed", err.Error(),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed token value from hwmux
	token, _, err := GetToken(r.client, &resp.Diagnostics, data.UserId.ValueString())
	if err != nil {
		return
	}

	err = updateTokenModelFromResponse(token, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the token model failed", err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// update token
	tokenSerializer, httpRes, err := r.client.UserApi.UserTokenCreate(context.Background(), data.UserId.ValueString()).Execute()

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating token "+data.ID.String(),
			"Could not update token, unexpected error: "+err.Error()+"\n"+BodyToString(&httpRes.Body),
		)
		return
	}

	// set model based on response
	err = updateTokenModelFromResponse(tokenSerializer, data, &resp.Diagnostics, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Updating the token model failed", err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing (we delete by creating a new one that invalidates the previous one)
	_, httpRes, err := r.client.UserApi.UserTokenCreate(context.Background(), data.UserId.ValueString()).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Token",
			"Could not delete token, unexpected error: "+BodyToString(&httpRes.Body),
		)
		return
	}
}

// Map response body to model and populate Computed attribute values
func updateTokenModelFromResponse(token *hwmux.Token, plan *TokenResourceModel, diagnostics *diag.Diagnostics, client *hwmux.APIClient) (err error) {
	// Map response body to schema and populate Computed attribute values
	plan.Token = types.StringValue(token.GetKey())
	plan.DateCreated = types.StringValue(token.GetCreated().Format(time.RFC850))

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	return nil
}
