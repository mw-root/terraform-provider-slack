// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccChannelDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccChannelDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.slack_channel.test_by_name", "id", "C085M89VBFH"),
					resource.TestCheckResourceAttr("data.slack_channel.test_by_id", "name", "test-channel"),
				),
			},
			{
				Config: providerConfig + testAccChannelDoesNotExistDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.slack_channel.does_not_exist", "id", "C085M89VBFH"),
				),
				ExpectError: regexp.MustCompile(`Unable to find channel`),
			},
		},
	})
}

const testAccChannelDataSourceConfig = `
data "slack_channel" "test_by_name" {
  name = "test-channel"
}
data "slack_channel" "test_by_id" {
  id = "C085M89VBFH"
}
`

const testAccChannelDoesNotExistDataSourceConfig = `
data "slack_channel" "does_not_exist" {
  name = "steve"
}
`
