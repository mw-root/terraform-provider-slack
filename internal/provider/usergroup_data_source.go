// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &UserGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &UserGroupDataSource{}
)

func NewUserGroupDataSource() datasource.DataSource {
	return &UserGroupDataSource{}
}

// UserGroupDataSource defines the data source implementation.
type UserGroupDataSource struct {
	client *slack.Client
}

// UserGroupDataSourceModel describes the data source data model.
type UserGroupDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Handle      types.String `tfsdk:"handle"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsExternal  types.Bool   `tfsdk:"is_external"`
}

func (d *UserGroupDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("handle"),
		),
	}
}

func (d *UserGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usergroup"
}

func (d *UserGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: `
Reads a slack User Group specified by handle or id.
### Required Permissions
- ` + "`usergroups:read`" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier for this User Group.",
				Optional:            true,
				Computed:            true,
			},
			"handle": schema.StringAttribute{
				MarkdownDescription: "The Slack mention handle of the User Group",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A name for the User Group.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A short description of the User Group.",
				Computed:            true,
			},
			"is_external": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the usergroup is an Admin of the current workspace.",
				Computed:            true,
			},
		},
	}
}

func (d *UserGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserGroupDataSourceModel
	var userGroup slack.UserGroup
	var err error

	client := d.client

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

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

	switch {
	case !data.Id.IsNull():
		userGroup, err = getUserGroupById(&userGroups, data.Id.ValueString())
	case !data.Handle.IsNull():
		userGroup, err = getUserGroupByHandle(&userGroups, data.Handle.ValueString())
	default:
		resp.Diagnostics.AddError("Provider Error", "One of ID or Handle needs to be provided.")
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find User Group, got error: %s", err))
		return
	}

	// Set data from API response.
	data.Id = types.StringValue(userGroup.ID)
	data.Handle = types.StringValue(userGroup.Handle)
	data.Name = types.StringValue(userGroup.Name)
	data.Description = types.StringValue(userGroup.Description)
	data.IsExternal = types.BoolValue(userGroup.IsExternal)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getUserGroupById(userGroups *[]slack.UserGroup, id string) (slack.UserGroup, error) {

	for _, each := range *userGroups {
		if each.ID == id {
			return each, nil
		}
	}

	return slack.UserGroup{}, fmt.Errorf("could not find user group %s", id)
}

func getUserGroupByHandle(userGroups *[]slack.UserGroup, handle string) (slack.UserGroup, error) {

	for _, each := range *userGroups {
		if each.Handle == handle {
			return each, nil
		}
	}

	return slack.UserGroup{}, fmt.Errorf("could not find user group %s", handle)
}
