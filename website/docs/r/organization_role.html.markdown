---
layout: "github"
page_title: "GitHub: github_organization_role Resource"
description: |-
  Manage a custom organization role.
---

# github_organization_role (Resource)

Manage a custom organization role.

~> **Note**: Custom organization roles are currently only available in GitHub Enterprise Cloud.

## Example Usage

```terraform
resource "github_organization_role" "example" {
  name      = "example"
  base_role = "read"

  permissions = [
    "read_organization_custom_org_role",
    "read_organization_custom_repo_role"
  ]
}
```

## Schema

### Required

- `name` (String) The name of the organization role.
- `permissions` (Set of String, Min: 1) The permissions included in this role.

### Optional

- `description` (String) The description of the organization role.
- `base_role` (String) The system role from which this role inherits permissions.

### Read-Only

- `role_id` (Number) The ID of the organization role.

## Import

A custom organization role can be imported using its ID.

```shell
terraform import github_organization_role.example 1234
```
