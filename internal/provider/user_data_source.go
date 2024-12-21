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
	_ datasource.DataSource              = &UserDataSource{}
	_ datasource.DataSourceWithConfigure = &UserDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	client *slack.Client
}

// UserDataSourceModel describes the data source data model.
type UserDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	RealName types.String `tfsdk:"real_name"`
	Deleted  types.Bool   `tfsdk:"deleted"`
	TimeZone types.String `tfsdk:"time_zone"`
	IsAdmin  types.Bool   `tfsdk:"is_admin"`
	IsBot    types.Bool   `tfsdk:"is_bot"`
}

func (d *UserDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `
Reads a slack user specified by name or id, and returns attributes.
### Required Permissions
- ` + "`users:read`" + `
`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The Slack handle of the user",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier for this workspace user. It is unique to the workspace containing the user.",
				Optional:            true,
				Computed:            true,
			},
			"real_name": schema.StringAttribute{
				MarkdownDescription: "The user's first and last name.",
				Computed:            true,
			},
			"deleted": schema.BoolAttribute{
				MarkdownDescription: "This user has been deactivated when the value of this field is `true`. Otherwise the value is `false`, or the field may not appear at all.",
				Computed:            true,
			},
			"time_zone": schema.StringAttribute{
				MarkdownDescription: "A human-readable string for the geographic timezone-related region this user has specified in their account.",
				Computed:            true,
			},
			"is_admin": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the user is an Admin of the current workspace.",
				Computed:            true,
			},
			"is_bot": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the user is actually a bot user. Bleep bloop. Note that Slackbot is special, so `is_bot` will be false for it.",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel
	var user slack.User
	var err error

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id.ValueString() != "" {
		user, err = getUserById(ctx, d.client, data.Id.ValueString())
	} else {
		user, err = getUserByName(ctx, d.client, data.Name.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find user, got error: %s", err))
		return
	}

	// Set data from API response.
	data.Id = types.StringValue(user.ID)
	data.Name = types.StringValue(user.Name)
	data.RealName = types.StringValue(user.RealName)
	data.Deleted = types.BoolValue(user.Deleted)
	data.TimeZone = types.StringValue(user.TZ)
	data.IsAdmin = types.BoolValue(user.IsAdmin)
	data.IsBot = types.BoolValue(user.IsBot)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getUserById(ctx context.Context, client *slack.Client, id string) (slack.User, error) {
	user, err := client.GetUserInfoContext(ctx, id)
	if err != nil {
		return slack.User{}, err
	}
	return *user, err
}

// This is basically the logic in slack.GetUsersContext.
// This is exploded here instead of using that method to ensure we're checking
// each returned page, potentially saving some API calls.
func getUserByName(ctx context.Context, client *slack.Client, name string) (slack.User, error) {

	tflog.Trace(ctx, "Requesting Page of Slack Users")

	var err interface{}
	err = nil

	page := client.GetUsersPaginated()

	for _, user := range page.Users {
		if user.Name == name {
			return user, nil
		}
	}

	for err == nil {
		page, err = page.Next(ctx)
		if err == nil {
			for _, user := range page.Users {
				if user.Name == name {
					return user, nil
				}
			}
		} else if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-time.After(rateLimitedError.RetryAfter):
				err = nil
			}
		}
	}

	return slack.User{}, fmt.Errorf("user: %s not found", name)

}
