---
layout: "github"
page_title: "GitHub: github_team"
description: |-
  Provides a GitHub team resource.
---

# github_team

Provides a GitHub team resource.

This resource allows you to add/remove teams from your organization. When applied,
a new team will be created. When destroyed, that team will be removed.

## Example Usage

```hcl
# Add a team to the organization
resource "github_team" "some_team" {
  name                 = "some-team"
  description          = "Some cool team"
  privacy              = "closed"
  notification_setting = "notifications_enabled"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the team.
* `description` - (Optional) A description of the team.
* `privacy` - (Optional) The level of privacy for the team. Must be one of `secret` or `closed`.
               Defaults to `secret`.
* `notification_setting` - (Optional) The notification setting for the team. Must be one of `notifications_enabled` or `notifications_disabled`.
                            Defaults to `notifications_disabled`
* `parent_team_id` - (Optional) The ID or slug of the parent team, if this is a nested team.
* `ldap_dn` - (Optional) The LDAP Distinguished Name of the group where membership will be synchronized. Only available in GitHub Enterprise Server.
* `create_default_maintainer` - (Optional) Adds a default maintainer to the team. Defaults to `false` and adds the creating user to the team when `true`.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the created team.
* `node_id` - The Node ID of the created team.
* `slug` - The slug of the created team, which may or may not differ from `name`,
  depending on whether `name` contains "URL-unsafe" characters.
  Useful when referencing the team in [`github_branch_protection`](/docs/providers/github/r/branch_protection.html).

## Import

GitHub Teams can be imported using the GitHub team ID or name e.g.

```
$ terraform import github_team.core 1234567
$ terraform import github_team.core Administrators
```
