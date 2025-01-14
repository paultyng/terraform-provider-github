---
layout: "github"
page_title: "GitHub: github_organization_role_users Data Source"
description: |-
  Lookup all users assigned to a custom organization role.
---

# github_organization_role_users (Data Source)

Lookup all users assigned to a custom organization role.

## Example Usage

```terraform
data "github_organization_role_users" "example" {
  role_id = 1234
}
```

## Schema

### Required

- `role_id` (Number) The ID of the organization role.

### Read-Only

- `users` (Set of Object, see [schema](#nested-schema-for-users)) Users assigned to the organization role.

## Nested Schema for `users`

### Read-Only

- `user_id` (Number) The ID of the user.
- `login` (String) The login for the GitHub user account.
