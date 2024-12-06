---
layout: "github"
page_title: "GitHub: github_project_column"
description: |-
  Creates and manages project columns for GitHub projects
---

# github_project_column

!> **Warning:** This resource no longer works as the [Projects (classic) REST API](https://docs.github.com/en/rest/projects/projects?apiVersion=2022-11-28) has been [removed](https://github.blog/changelog/2024-05-23-sunset-notice-projects-classic/) and as such has been deprecated. It will be removed in a future release.

This resource allows you to create and manage columns for GitHub projects.

## Example Usage

```hcl
resource "github_organization_project" "project" {
  name = "A Organization Project"
  body = "This is an organization project."
}

resource "github_project_column" "column" {
  project_id = github_organization_project.project.id
  name       = "a column"
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) The ID of an existing project that the column will be created in.

* `name` - (Required) The name of the column.
