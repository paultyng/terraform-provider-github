---
layout: "github"
page_title: "GitHub: github_organization_project"
description: |-
  Creates and manages projects for GitHub organizations
---

# github_organization_project

!> **Warning:** This resource no longer works as the [Projects (classic) REST API](https://docs.github.com/en/rest/projects/projects?apiVersion=2022-11-28) has been [removed](https://github.blog/changelog/2024-05-23-sunset-notice-projects-classic/) and as such has been deprecated. It will be removed in a future release.

This resource allows you to create and manage projects for GitHub organization.

## Example Usage

```hcl
resource "github_organization_project" "project" {
  name = "A Organization Project"
  body = "This is a organization project."
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the project.

* `body` - (Optional) The body of the project.

## Attributes Reference

The following additional attributes are exported:

* `url` - URL of the project
