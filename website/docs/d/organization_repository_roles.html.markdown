---
layout: "github"
page_title: "GitHub: github_organization_repository_roles Data Source"
description: |-
  Lookup all custom repository roles in an organization.
---

# github_organization_repository_roles (Data Source)

Lookup all custom repository roles in an organization.

~> **Note**: Custom organization repository roles are currently only available in GitHub Enterprise Cloud.

## Example Usage

```terraform
data "github_organization_repository_roles" "example" {
}
```

## Schema

### Read-Only

- `roles` (Set of Object, see [schema](#nested-schema-for-roles)) Available organization repository roles.

## Nested Schema for `roles`

### Read-Only

- `role_id` (Number) The ID of the organization repository role.
- `name` (String) The name of the organization repository role.
- `description` (String) The description of the organization repository role.
- `base_role` (String) The system role from which this role inherits permissions.
- `permissions` (Set of String) The permissions included in this role.
