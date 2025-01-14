---
layout: "github"
page_title: "GitHub: github_organization_role_user Resource"
description: |-
  Manage an association between an organization role and a user.
---

# github_organization_role_user (Resource)

Manage an association between an organization role and a user.

## Example Usage

```terraform
resource "github_organization_role_user" "example" {
  role_id = 1234
  login   = "example-user"
}
```

## Schema

### Required

- `role_id` (Number) The ID of the organization role.
- `login` (String) The login for the GitHub user account.

## Import

An organization role user association can be imported using the role ID and the user login separated by a `:`.

```shell
terraform import github_organization_role_team.example "1234:example-user"
```
