---
layout: "github"
page_title: "GitHub: github_organization_role_teams Data Source"
description: |-
  Lookup all teams assigned to a custom organization role.
---

# github_organization_role_teams (Data Source)

Lookup all teams assigned to a custom organization role.

## Example Usage

```terraform
data "github_organization_role_teams" "example" {
  role_id = 1234
}
```

## Schema

### Required

- `role_id` (Number) The ID of the organization role.

### Read-Only

- `teams` (Set of Object, see [schema](#nested-schema-for-teams)) Teams assigned to the organization role.

## Nested Schema for `teams`

### Read-Only

- `team_id` (Number) The ID of the team.
- `slug` (String) The Slug of the team name.
- `name` (String) The name of the team.
- `permission` (String) The permission that the team will have for its repositories.
