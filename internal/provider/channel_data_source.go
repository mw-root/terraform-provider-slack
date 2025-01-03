// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &ChannelDataSource{}
	_ datasource.DataSourceWithConfigure = &ChannelDataSource{}
)

func NewChannelDataSource() datasource.DataSource {
	return &ChannelDataSource{}
}

// ChannelDataSource defines the data source implementation.
type ChannelDataSource struct {
	client *slack.Client
}

// ChannelDataSourceModel describes the data source data model.
type ChannelDataSourceModel struct {
	Name            types.String `tfsdk:"name"`
	Id              types.String `tfsdk:"id"`
	IncludeArchived types.Bool   `tfsdk:"include_archived"`
	Topic           types.String `tfsdk:"topic"`
	Description     types.String `tfsdk:"description"`
}

func (d *ChannelDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *ChannelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_channel"
}

func (d *ChannelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `
Reads a slack channel specified by name or id, and returns attributes.
### Required Permissions
- ` + "`channel:read`" + `
`,

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the channel",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The Channel ID",
				Optional:            true,
				Computed:            true,
			},
			"include_archived": schema.BoolAttribute{
				MarkdownDescription: "Set true to include archived channels.",
				Optional:            true,
			},
			"topic": schema.StringAttribute{
				MarkdownDescription: "The Channel's configured topic.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The Channel's configured description.",
				Computed:            true,
			},
		},
	}
}

func (d *ChannelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func getChannelById(ctx context.Context, client *slack.Client, id string) (slack.Channel, error) {
	channel, err := client.GetConversationInfoContext(
		ctx,
		&slack.GetConversationInfoInput{
			ChannelID:         id,
			IncludeLocale:     false,
			IncludeNumMembers: false,
		},
	)
	if err != nil {
		return slack.Channel{}, err
	}
	return *channel, err

}

func getChannelByName(ctx context.Context, client *slack.Client, name string, excludeArchived bool) (slack.Channel, error) {

	var err error
	var cursor string
	var nextCursor string
	var channels []slack.Channel

	err = nil
	cursor = ""

	for err == nil {

		tflog.Trace(ctx, fmt.Sprintf("Exclude Archived: %t", excludeArchived))
		params := &slack.GetConversationsParameters{
			ExcludeArchived: excludeArchived,
			Cursor:          cursor,
		}

		tflog.Trace(ctx, "Next Cursor: "+cursor)

		channels, nextCursor, err = client.GetConversationsContext(
			ctx,
			params,
		)

		if err == nil {
			tflog.Trace(ctx, "Searching Page for: "+name)

			for _, channel := range channels {
				if channel.Name == name {
					tflog.Trace(ctx, "Found channel: "+name)
					return channel, nil
				}
			}
			tflog.Trace(ctx, "Channel not found in page.")

			if nextCursor == "" {
				tflog.Trace(ctx, "We have reached the last page of results and have not found this channel.")
				return slack.Channel{}, fmt.Errorf("channel_not_found")
			}
			cursor = nextCursor
			continue

		} else if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {

			tflog.Trace(ctx, rateLimitedError.Error())
			select {
			case <-ctx.Done():
				err = ctx.Err()
				tflog.Error(ctx, "Context is Done. "+err.Error())
			case <-time.After(rateLimitedError.RetryAfter):
				err = nil
			}
		}

	}

	return slack.Channel{}, fmt.Errorf("error listing channels: %s", err.Error())
}

func (d *ChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ChannelDataSourceModel
	var channel slack.Channel
	var err error

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.ValueString() != "" {
		channel, err = getChannelById(ctx, d.client, data.Id.ValueString())

	} else {
		channel, err = getChannelByName(ctx, d.client, data.Name.ValueString(), !data.IncludeArchived.ValueBool())
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find channel, got error: %s", err))
	}

	// Set data from API response.
	data.Id = types.StringValue(channel.ID)
	data.Name = types.StringValue(channel.Name)
	data.Description = types.StringValue(channel.Purpose.Value)
	data.Topic = types.StringValue(channel.Topic.Value)

	data.IncludeArchived = types.BoolValue(data.IncludeArchived.ValueBool())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
