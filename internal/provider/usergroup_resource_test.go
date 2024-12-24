// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testUserGroupResourceName string = "test-usergroup-" + testResourceNameSuffix
var testUserGroupResourceDescription string = "Test Description " + testResourceNameSuffix
var testUserGroupResourceHandle string = "test-handle-" + testResourceNameSuffix

func TestUserGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "slack_usergroup" "test" {
  name = "` + testUserGroupResourceName + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("slack_usergroup.test", "name", testUserGroupResourceName),
					resource.TestCheckResourceAttr("slack_usergroup.test", "description", ""),
					resource.TestCheckResourceAttr("slack_usergroup.test", "handle", ""),
				),
			},
			// ImportState testing
			{
				ResourceName:      "slack_usergroup.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "slack_usergroup" "test" {
  name         = "` + testUserGroupResourceName + `"
  description  = "` + testUserGroupResourceDescription + `"
  handle       = "` + testUserGroupResourceHandle + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("slack_usergroup.test", "name", testUserGroupResourceName),
					resource.TestCheckResourceAttr("slack_usergroup.test", "description", testUserGroupResourceDescription),
					resource.TestCheckResourceAttr("slack_usergroup.test", "handle", testUserGroupResourceHandle),
				),
			},
			// Test Removal of Topic and Desc values
			{
				Config: providerConfig + `
resource "slack_usergroup" "test" {
  name   = "` + testUserGroupResourceName + `"
  handle = "` + testUserGroupResourceHandle + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("slack_usergroup.test", "name", testUserGroupResourceName),
					resource.TestCheckResourceAttr("slack_usergroup.test", "description", ""),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
