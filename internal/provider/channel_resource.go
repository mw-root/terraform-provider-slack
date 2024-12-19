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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ChannelResource{}
var _ resource.ResourceWithImportState = &ChannelResource{}

func NewChannelResource() resource.Resource {
	return &ChannelResource{}
}

// ChannelResource defines the resource implementation.
type ChannelResource struct {
	client *slack.Client
}

// ChannelResourceModel describes the resource data model.
type ChannelResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Id          types.String `tfsdk:"id"`
	IsPrivate   types.Bool   `tfsdk:"is_private"`
	Topic       types.String `tfsdk:"topic"`
	Description types.String `tfsdk:"description"`
}

func (r *ChannelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel"
}

func (r *ChannelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `
Creates a public or private slack channel.
### Required Permissions
` + "- `channels:manage`" + `
`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the channel to create.",
				Required:            true,
			},
			"is_private": schema.BoolAttribute{
				MarkdownDescription: "Create a private channel instead of a public one.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "The Channel's topic.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The Channel's description.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Channel identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ChannelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ChannelResourceModel
	client := r.client

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := slack.CreateConversationParams{
		ChannelName: data.Name.ValueString(),
		IsPrivate:   data.IsPrivate.ValueBool(),
	}

	created, err := client.CreateConversationContext(
		ctx,
		params,
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create channel: %s, got error: %s", params.ChannelName, err))
		return
	}

	if data.Description.ValueString() != "" {
		tflog.Trace(ctx, "Setting channel description")

		_, err := client.SetPurposeOfConversationContext(
			ctx, created.ID, data.Description.ValueString(),
		)

		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set channel description, got error: %s", err))
			return
		}
	}

	if data.Topic.ValueString() != "" {
		tflog.Trace(ctx, "Setting channel description")

		_, err := client.SetTopicOfConversationContext(ctx, created.ID, data.Topic.ValueString())

		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set channel description, got error: %s", err))
			return
		}
	}

	channel, err := getChannelById(ctx, client, created.ID)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read channel, got error: %s", err))
		return
	}

	data.Id = types.StringValue(channel.ID)
	data.Name = types.StringValue(channel.Name)
	data.IsPrivate = types.BoolValue(channel.IsPrivate)
	data.Topic = types.StringValue(channel.Topic.Value)
	data.Description = types.StringValue(channel.Purpose.Value)

	tflog.Trace(ctx, "Created a slack channel")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ChannelResourceModel
	client := r.client

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := getChannelById(ctx, client, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read channel, got error: %s", err))
		return
	}

	data.Id = types.StringValue(channel.ID)
	data.Name = types.StringValue(channel.Name)
	data.IsPrivate = types.BoolValue(channel.IsPrivate)
	data.Topic = types.StringValue(channel.Topic.Value)
	data.Description = types.StringValue(channel.Purpose.Value)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ChannelResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	client := r.client

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		tflog.Trace(ctx, "Updating Channel Name")

		_, err := client.RenameConversationContext(
			ctx, state.Id.ValueString(), plan.Name.ValueString(),
		)

		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update channel name, got error: %s", err))
			return
		}
	}

	if !plan.Description.Equal(state.Description) {
		tflog.Trace(ctx, "Updating Channel Description")

		_, err := client.SetPurposeOfConversationContext(
			ctx, state.Id.ValueString(), plan.Description.ValueString(),
		)

		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update channel description, got error: %s", err))
			return
		}
	}

	if !plan.Topic.Equal(state.Topic) {
		tflog.Trace(ctx, "Updating Channel Topic")

		_, err := client.SetTopicOfConversationContext(
			ctx, state.Id.ValueString(), plan.Topic.ValueString(),
		)

		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update channel topic, got error: %s", err))
			return
		}
	}

	channel, err := getChannelById(ctx, client, state.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read channel, got error: %s", err))
		return
	}

	plan.Name = types.StringValue(channel.Name)
	plan.IsPrivate = types.BoolValue(channel.IsPrivate)
	plan.Topic = types.StringValue(channel.Topic.Value)
	plan.Description = types.StringValue(channel.Purpose.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ChannelResourceModel
	client := r.client

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := client.ArchiveConversationContext(
		ctx, data.Id.ValueString(),
	)
	if err != nil {
		if err.Error() == "channel_not_found" {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to archive channel, got error: %s", err))
		return
	}

}

func (r *ChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
