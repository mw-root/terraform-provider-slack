// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/slack-go/slack"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure SlackProvider satisfies various provider interfaces.
var _ provider.Provider = &SlackProvider{}

// SlackProvider defines the provider implementation.
type SlackProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SlackProviderModel describes the provider data model.
type SlackProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func (p *SlackProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "slack"
	resp.Version = p.version
}

func (p *SlackProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
A provider to interact with and manage Slack resources.

A slack bot and its OAuth token is required to make use of this provider. 
Each resource and data source will document the permissions (Bot Token Scopes) required to perform that operation.
`,
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "Slack API Token. This can also be set by configuring the `SLACK_TOKEN` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *SlackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SlackProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("SLACK_TOKEN")

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}
	client := slack.New(token)
	_, err := client.AuthTest()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Configure Slack Client",
			"An unexpected error occurred when testing the slack API. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Slack Client Error: "+err.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SlackProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// NewExampleResource,
	}
}

func (p *SlackProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource,
		NewChannelDataSource,
	}
}

func (p *SlackProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SlackProvider{
			version: version,
		}
	}
}
