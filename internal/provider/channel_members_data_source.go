// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ChannelMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &ChannelMembersDataSource{}
)

func NewChannelMembersDataSource() datasource.DataSource {
	return &ChannelMembersDataSource{}
}

// ChannelMembersDataSource defines the data source implementation.
type ChannelMembersDataSource struct {
	client *slack.Client
}

// ChannelMembersDataSourceModel describes the data source data model.
type ChannelMembersDataSourceModel struct {
	Id      types.String `tfsdk:"id"`
	Members types.Set    `tfsdk:"members"`
}

func (d *ChannelMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel_members"
}

func (d *ChannelMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `
Gets the Slack IDs of a given channel's members.
### Required Permissions
- ` + "`channels:read`" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ChannelMembers ID",
				Optional:            true,
				Computed:            true,
			},
			"members": schema.SetAttribute{
				MarkdownDescription: "Set of channel member's Slack IDs.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *ChannelMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*slack.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *slack.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ChannelMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ChannelMembersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var allMembers []string

	members, next, err := d.client.GetUsersInConversationContext(
		ctx,
		&slack.GetUsersInConversationParameters{
			ChannelID: data.Id.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find channel members, got error: %s", err))
	}
	allMembers = append(allMembers, members...)

	for next != "" {
		members, next, err = d.client.GetUsersInConversationContext(
			ctx,
			&slack.GetUsersInConversationParameters{
				ChannelID: data.Id.ValueString(),
				Cursor:    next,
			},
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find channel members, got error: %s", err))
		}
		allMembers = append(allMembers, members...)
	}

	var diags diag.Diagnostics

	// Set data from API response.
	data.Members, diags = types.SetValueFrom(ctx, types.StringType, allMembers)

	resp.Diagnostics.Append(diags...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
