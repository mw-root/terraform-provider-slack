// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testUserGroupId = "S085R0X76CX"
const testUserGroupHandle = "test-group"

func TestAccUserGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccUserGroupDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.slack_usergroup.test_by_id", "handle", testUserGroupHandle),
					resource.TestCheckResourceAttr("data.slack_usergroup.test_by_handle", "id", testUserGroupId),
				),
			},
			{
				Config: providerConfig + testAccUserGroupDoesNotExistDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.slack_user.does_not_exist", "id", ""),
				),
				ExpectError: regexp.MustCompile(`Unable to find User Group`),
			},
		},
	})
}

const testAccUserGroupDataSourceConfig = `
data "slack_usergroup" "test_by_id" {
  id = "` + testUserGroupId + `"
}
data "slack_usergroup" "test_by_handle" {
  handle = "` + testUserGroupHandle + `"
}
`

const testAccUserGroupDoesNotExistDataSourceConfig = `
data "slack_usergroup" "does_not_exist" {
  handle = "abcde"
}
`
