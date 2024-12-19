// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
	"time"

	"math/rand"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var testResourceNameSuffix string = randSeq(6)
var testChannelName string = "test-channel-" + testResourceNameSuffix
var testChannelDescription string = "Test Description " + testResourceNameSuffix
var testChannelTopic string = "Test Topic " + testResourceNameSuffix

func TestChannelResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "slack_channel" "test" {
  name = "` + testChannelName + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("slack_channel.test", "name", testChannelName),
					resource.TestCheckResourceAttr("slack_channel.test", "description", ""),
					resource.TestCheckResourceAttr("slack_channel.test", "topic", ""),
				),
			},
			// ImportState testing
			{
				ResourceName:      "slack_channel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "slack_channel" "test" {
  name        = "` + testChannelName + `"
  description = "` + testChannelDescription + `"
  topic       = "` + testChannelTopic + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("slack_channel.test", "name", testChannelName),
					resource.TestCheckResourceAttr("slack_channel.test", "description", testChannelDescription),
					resource.TestCheckResourceAttr("slack_channel.test", "topic", testChannelTopic),
				),
			},
			// Test Removal of Topic and Desc values
			{
				Config: providerConfig + `
resource "slack_channel" "test" {
  name = "` + testChannelName + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("slack_channel.test", "name", testChannelName),
					resource.TestCheckResourceAttr("slack_channel.test", "description", ""),
					resource.TestCheckResourceAttr("slack_channel.test", "topic", ""),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
