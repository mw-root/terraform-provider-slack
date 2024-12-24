---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "slack_usergroup Resource - slack"
subcategory: ""
description: |-
  Creates a Slack User Group.
  Required Permissions
  usergroups:write
---

# slack_usergroup (Resource)

Creates a Slack User Group.
### Required Permissions
- `usergroups:write`

## Example Usage

```terraform
resource "slack_usergroup" "group" {
  handle      = "my-group"
  name        = "My Group"
  description = "User group for stuff and/or things"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A name for the User Group. Must be unique among User Groups.

### Optional

- `description` (String) A short description of the User Group.
- `handle` (String) A mention handle. Must be unique among channels, users and User Groups.

### Read-Only

- `id` (String) Identifier for this User Group.