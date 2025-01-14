---
layout: "github"
page_title: "GitHub: github_organization_roles Data Source"
description: |-
  Lookup all custom roles in an organization.
---

# github_organization_roles (Data Source)

Lookup all custom roles in an organization.

## Example Usage

```terraform
data "github_organization_roles" "example" {
}
```

## Schema

### Read-Only

- `roles` (Set of Object, see [schema](#nested-schema-for-roles)) Available organization roles.

## Nested Schema for `roles`

### Read-Only

- `role_id` (Number) The ID of the organization role.
- `name` (String) The name of the organization role.
- `description` (String) The description of the organization role.
- `source` (String) The source of this role; one of `Predefined`, `Organization`, or `Enterprise`.
- `base_role` (String) The system role from which this role inherits permissions.
- `permissions` (Set of String) The permissions included in this role.
