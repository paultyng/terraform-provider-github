---
layout: "github"
page_title: "GitHub: github_organization_repository_role Data Source"
description: |-
  Lookup a custom organization repository role.
---

# github_organization_repository_role (Data Source)

Lookup a custom organization repository role.

~> **Note**: Custom organization repository roles are currently only available in GitHub Enterprise Cloud.

## Example Usage

```terraform
data "github_organization_repository_role" "example" {
  role_id = 1234
}
```

## Schema

### Required

- `role_id` (Number) The ID of the organization repository role.

### Read-Only

- `name` (String) The name of the organization repository role.
- `description` (String) The description of the organization repository role.
- `base_role` (String) The system role from which this role inherits permissions.
- `permissions` (Set of String) The permissions included in this role.
