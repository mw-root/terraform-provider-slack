// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testDataSourceChannelMembersChannelId = "C086QLHRNV6"
const testDataSourceChannelMembersChannelMemberId = "U085RJKA41X"

func TestAccChannelMembersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccChannelMembersDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.slack_channel_members.test", "id", testDataSourceChannelMembersChannelId),
					resource.TestCheckTypeSetElemAttr("data.slack_channel_members.test", "members.*", testDataSourceChannelMembersChannelMemberId),
				),
			},
			{
				Config: providerConfig + testAccChannelMembersDataSourceConfigChannelDoesNotExist,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.slack_channel.does_not_exist", "id", ""),
				),
				ExpectError: regexp.MustCompile(`Unable to find channel`),
			},
		},
	})
}

const testAccChannelMembersDataSourceConfig = `
data "slack_channel_members" "test" {
  id = "` + testDataSourceChannelMembersChannelId + `"
}
`

const testAccChannelMembersDataSourceConfigChannelDoesNotExist = `
data "slack_channel_members" "does_not_exist" {
  id = "CDOESNOTEXIST"
}
`
