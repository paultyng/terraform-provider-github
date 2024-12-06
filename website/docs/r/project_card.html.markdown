---
layout: "github"
page_title: "GitHub: github_project_card"
description: |-
  Creates and manages project cards for GitHub projects
---

# github_project_card

!> **Warning:** This resource no longer works as the [Projects (classic) REST API](https://docs.github.com/en/rest/projects/projects?apiVersion=2022-11-28) has been [removed](https://github.blog/changelog/2024-05-23-sunset-notice-projects-classic/) and as such has been deprecated. It will be removed in a future release.

This resource allows you to create and manage cards for GitHub projects.

## Example Usage

```hcl
resource "github_organization_project" "project" {
  name = "An Organization Project"
  body = "This is an organization project."
}

resource "github_project_column" "column" {
  project_id = github_organization_project.project.id
  name       = "Backlog"
}

resource "github_project_card" "card" {
  column_id = github_project_column.column.column_id
  note        = "## Unaccepted 👇"
}
```
## Example Usage adding an Issue to a Project

```hcl
resource "github_repository" "test" {
  name = "myrepo"
  has_projects = true
  has_issues   = true
}

resource "github_issue" "test" {
  repository       = github_repository.test.id
  title            = "Test issue title"
  body             = "Test issue body"
}

resource "github_repository_project" "test" {
  name            = "test"
  repository      = github_repository.test.name
  body            = "this is a test project"
}

resource "github_project_column" "test" {
  project_id = github_repository_project.test.id
  name       = "Backlog"
}

resource "github_project_card" "test" {
  column_id    = github_project_column.test.column_id
  content_id   = github_issue.test.issue_id
  content_type = "Issue"
}
```

## Argument Reference

The following arguments are supported:

* `column_id` - (Required) The ID of the card.

* `note` - (Optional) The note contents of the card. Markdown supported.

* `content_id` - (Optional) [`github_issue.issue_id`](issue.html#argument-reference).

* `content_type` - (Optional) Must be either `Issue` or `PullRequest`

**Remarks:** You must either set the `note` attribute or both `content_id` and `content_type`.
See [note example](#example-usage) or [issue example](#example-usage-adding-an-issue-to-a-project) for more information.

## Import

A GitHub Project Card can be imported using its [Card ID](https://developer.github.com/v3/projects/cards/#get-a-project-card):

```
$ terraform import github_project_card.card 01234567
```
