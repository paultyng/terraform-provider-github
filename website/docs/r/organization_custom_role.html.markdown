---
layout: "github"
page_title: "GitHub: github_organization_custom_role"
description: |-
  Creates and manages a custom role in a GitHub Organization for use in repositories.
---

# github\_organization\_custom\_role

~> **Note:** This resource is deprecated, please use the `github_organizationrepository_role` resource instead.

This resource allows you to create and manage custom roles in a GitHub Organization for use in repositories.

~> Note: Custom roles are currently only available in GitHub Enterprise Cloud.

## Example Usage

```hcl
resource "github_organization_custom_role" "example" {
  name = "example"
  description = "Example custom role that uses the read role as its base"
  base_role = "read"
  permissions = [
    "add_assignee",
    "add_label",
    "bypass_branch_protection",
    "close_issue",
    "close_pull_request",
    "mark_as_duplicate",
    "create_tag",
    "delete_issue",
    "delete_tag",
    "manage_deploy_keys",
    "push_protected_branch",
    "read_code_scanning",
    "reopen_issue",
    "reopen_pull_request",
    "request_pr_review",
    "resolve_dependabot_alerts",
    "resolve_secret_scanning_alerts",
    "view_secret_scanning_alerts",
    "write_code_scanning"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the custom role.
* `description` - (Optional) The description for the custom role.
* `base_role` - (Required) The system role from which the role inherits permissions. Can be one of: `read`, `triage`, `write`, or `maintain`.
* `permissions` - (Required) A list of additional permissions included in this role. Must have a minimum of 1 additional permission. The list of available permissions can be found using the [list repository fine-grained permissions for an organization](https://docs.github.com/en/enterprise-cloud@latest/rest/orgs/custom-roles?apiVersion=2022-11-28#list-repository-fine-grained-permissions-for-an-organization) API.

## Attributes Reference

The following additional attributes are exported:

* `id` - The ID of the custom role.

## Import

Custom roles can be imported using the `id` of the role.
The `id` of the custom role can be found using the [list custom roles in an organization](https://docs.github.com/en/enterprise-cloud@latest/rest/orgs/custom-roles#list-custom-repository-roles-in-an-organization) API.

```
$ terraform import github_organization_custom_role.example 1234
```
