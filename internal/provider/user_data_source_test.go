// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testUserName = "mike.weeks"
const testUserId = "U06F3KHU2J2"

func TestAccUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccUserDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.slack_user.test_by_name", "id", testUserId),
					resource.TestCheckResourceAttr("data.slack_user.test_by_id", "name", testUserName),
				),
			},
			{
				Config: providerConfig + testAccUserDoesNotExistDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.slack_user.does_not_exist", "id", ""),
				),
				ExpectError: regexp.MustCompile(`Unable to find user`),
			},
		},
	})
}

const testAccUserDataSourceConfig = `
data "slack_user" "test_by_name" {
  name = "` + testUserName + `"
}
data "slack_user" "test_by_id" {
  id = "` + testUserId + `"
}
`

const testAccUserDoesNotExistDataSourceConfig = `
data "slack_user" "does_not_exist" {
  name = "abcde"
}
`
