---
layout: "github"
page_title: "GitHub: github_external_groups"
description: |-
  Retrieve external groups belonging to an organization.
---

# github\_external\_groups

Use this data source to retrieve external groups belonging to an organization.

## Example Usage

```hcl
data "github_external_groups" "example_external_groups" {}

locals {
  local_groups = "${data.github_external_groups.example_external_groups}"
}

output "groups" {
  value = local.local_groups
}
```

## Argument Reference

N/A. This resource will retrieve all the external groups belonging to an organization.

## Attributes Reference

 * `external_groups` - an array of external groups belonging to the organization. Each group consists of the fields documented below.

___


 * `group_id` - the ID of the group.
 * `group_name` - the name of the group.
 * `updated_at` - the date the group was last updated.

