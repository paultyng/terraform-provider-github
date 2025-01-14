---
layout: "github"
page_title: "GitHub: github_organization_security_managers"
description: |-
  Get the security managers for an organization.
---

# github_organization_security_managers

Use this data source to retrieve the security managers for an organization.

## Example Usage

```hcl
data "github_organization_security_managers" "test" {}
```

## Attributes Reference

 * `teams` - An list of GitHub teams.  Each `team` block consists of the fields documented below.

___

The `team` block consists of:

 * `id` - Unique identifier of the team.
 * `slug` - Name based identifier of the team.
 * `name` - Name of the team.
 * `permission` - Permission that the team will have for its repositories.
