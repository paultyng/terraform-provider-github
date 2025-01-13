---
layout: "github"
page_title: "GitHub: github_workflow_repository_permissions"
description: |-
  Enables and manages Workflow permissions for a GitHub repository
---

# github_workflow_repository_permissions

This resource allows you to manage GitHub Workflow permissions for a given repository.
You must have admin access to a repository to use this resource.

## Example Usage

```hcl
resource "github_repository" "example" {
  name = "my-repository"
}

resource "github_workflow_repository_permissions" "test" {
  default_workflow_permissions = "read"
  can_approve_pull_request_reviews = true
  repository = github_repository.example.name
}
```

## Argument Reference

The following arguments are supported:

* `repository`                       - (Required) The GitHub repository
* `default_workflow_permissions`     - (Optional) The default workflow permissions granted to the GITHUB_TOKEN when running workflows. Can be one of: `read` or `write`.
* `can_approve_pull_request_reviews` - (Optional) Whether GitHub Actions can approve pull requests. Enabling this can be a security risk.

## Import

This resource can be imported using the name of the GitHub repository:

```
$ terraform import github_workflow_repository_permissions.test my-repository
```
