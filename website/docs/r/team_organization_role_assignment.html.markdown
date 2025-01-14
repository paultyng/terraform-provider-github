---
layout: "github"
page_title: "GitHub: github_team_organization_role_assignment"
description: |-
  Manages the associations between teams and organization roles.
---

# github_team_organization_role_assignment

This resource manages relationships between teams and organization roles
in your GitHub organization. This works on predefined roles, and custom roles, where the latter is a Enterprise feature.

Creating this resource assigns the role to a team.

The organization role and team must both belong to the same organization
on GitHub.

## Example Usage

```hcl
resource "github_team" "test-team" {
  name     = "test-team"
}

resource "github_team_organization_role_assignment" "test-team-role-assignment" {
  team_slug = github_team.test-team.slug
  role_id   = "8132" # all_repo_read (predefined)
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) The GitHub team id or the GitHub team slug
* `role_id` - (Required) The GitHub Organization Role id

## Import

GitHub Team Organization Role Assignment can be imported using an ID made up of `team_id:role_id` or `team_slug:role_id`, e.g.

```
$ terraform import github_team_organization_role_assignment.role_assignment 1234567:8132
$ terraform import github_team_organization_role_assignment.role_assignment test-team:8132
```
