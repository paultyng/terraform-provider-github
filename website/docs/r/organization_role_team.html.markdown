---
layout: "github"
page_title: "GitHub: github_organization_role_team Resource"
description: |-
  Manage an association between an organization role and a team.
---

# github_organization_role_team (Resource)

Manage an association between an organization role and a team.

## Example Usage

```terraform
resource "github_organization_role_team" "example" {
  role_id   = 1234
  team_slug = "example-team"
}
```

## Schema

### Required

- `role_id` (Number) The ID of the organization role.
- `team_slug` (String) The slug of the team name.

## Import

An organization role team association can be imported using the role ID and the team slug separated by a `:`.

```shell
terraform import github_organization_role_team.example "1234:example-team"
```
