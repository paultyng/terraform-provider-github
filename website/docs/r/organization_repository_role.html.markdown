---
layout: "github"
page_title: "GitHub: github_organization_repository_role Resource"
description: |-
  Manage a custom organization repository role.
---

# github_organization_repository_role (Resource)

Manage a custom organization repository role.

~> **Note**: Custom organization repository roles are currently only available in GitHub Enterprise Cloud.

## Example Usage

```terraform
resource "github_organization_repository_role" "example" {
  name      = "example"
  base_role = "read"

  permissions = [
    "add_assignee",
    "add_label"
  ]
}
```

## Schema

### Required

- `name` (String) The name of the organization repository role.
- `base_role` (String) The system role from which this role inherits permissions.
- `permissions` (Set of String, Min: 1) The permissions included in this role.

### Optional

- `description` (String) The description of the organization repository role.

### Read-Only

- `role_id` (Number) The ID of the organization repository role.

## Import

A custom organization repository role can be imported using its ID.

```shell
terraform import github_organization_repository_role.example 1234
```
