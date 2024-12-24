// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserGroupResource{}
var _ resource.ResourceWithImportState = &UserGroupResource{}

func NewUserGroupResource() resource.Resource {
	return &UserGroupResource{}
}

// UserGroupResource defines the resource implementation.
type UserGroupResource struct {
	client *slack.Client
}

// UserGroupResourceModel describes the resource data model.
type UserGroupResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Handle      types.String `tfsdk:"handle"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *UserGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usergroup"
}

func (r *UserGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `
Creates a Slack User Group.
### Required Permissions
` + "- `usergroups:write`" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier for this User Group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A name for the User Group. Must be unique among User Groups.",
				Required:            true,
			},
			"handle": schema.StringAttribute{
				MarkdownDescription: "A mention handle. Must be unique among channels, users and User Groups.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(
							ctx context.Context,
							sr planmodifier.StringRequest,
							rrifr *stringplanmodifier.RequiresReplaceIfFuncResponse,
						) {
							rrifr.RequiresReplace = !(sr.StateValue.ValueString() == "") && sr.PlanValue.ValueString() == ""
						},
						"Handle cannot be removed once it is set.",
						"Handle cannot be removed once it is set.",
					),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A short description of the User Group.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
		},
	}
}

func (r *UserGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*slack.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *slack.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *UserGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserGroupResourceModel
	client := r.client

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := slack.UserGroup{
		Name:        data.Name.ValueString(),
		Handle:      data.Handle.ValueString(),
		Description: data.Description.ValueString(),
	}
	userGroup, err := client.CreateUserGroupContext(ctx, params)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create User Group, got error: %s", err))
		return
	}

	data.Id = types.StringValue(userGroup.ID)
	data.Description = types.StringValue(userGroup.Description)
	data.Name = types.StringValue(userGroup.Name)
	data.Handle = types.StringValue(userGroup.Handle)

	tflog.Trace(ctx, "Created a slack User Group")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserGroupResourceModel
	client := r.client

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userGroups, err := client.GetUserGroupsContext(
		ctx,
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find User Group, got error: %s", err))
		return
	}

	userGroup, err := getUserGroupById(&userGroups, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find User Group, got error: %s", err))
		return
	}

	data.Name = types.StringValue(userGroup.Name)
	data.Description = types.StringValue(userGroup.Description)
	data.Handle = types.StringValue(userGroup.Handle)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UserGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	client := r.client

	if resp.Diagnostics.HasError() {
		return
	}

	params := []slack.UpdateUserGroupsOption{
		slack.UpdateUserGroupsOptionName(plan.Name.ValueString()),
		slack.UpdateUserGroupsOptionHandle(plan.Handle.ValueString()),
		slack.UpdateUserGroupsOptionDescription(plan.Description.ValueStringPointer()),
	}

	userGroup, err := client.UpdateUserGroupContext(ctx, plan.Id.ValueString(), params...)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to Update User Group, got error: %s", err))
		return
	}

	plan.Name = types.StringValue(userGroup.Name)
	plan.Description = types.StringValue(userGroup.Description)
	plan.Handle = types.StringValue(userGroup.Handle)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserGroupResourceModel
	client := r.client

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := client.DisableUserGroupContext(
		ctx, data.Id.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable User Group, got error: %s", err))
		return
	}

}

func (r *UserGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
