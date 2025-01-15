---
layout: "github"
page_title: "GitHub: github_organization_repositories"
description: |-
  Read details of all repositories of an organization.
---

# github\_organization\_repositories

Use this data source to retrieve all repositories of the organization.

## Example Usage

To retrieve *all* repositories of the organization:

```hcl
data "github_organization_repositories" "all" {}
```

## Attributes Reference

* `repository` - An Array of GitHub repositories.  Each `repository` block consists of the fields documented below.
___

The `repository` block consists of:

 * `repo_id` - GitHub ID for the repository.
 * `node_id` - The Node ID of the repository.
 * `name` - The name of the repository.
 * `archived` - Whether the repository is archived.
 * `visibility` - Whether the repository is public, private or internal.
