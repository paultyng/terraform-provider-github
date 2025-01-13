---
layout: "github"
page_title: "GitHub: github_app_installation_repository"
description: |-
  Manages the associations between app installations and repositories.
---

# github_app_installation_repository

~> **Note**: This resource is not compatible with the GitHub App Installation authentication method.

This resource manages relationships between app installations and repositories
in your GitHub organization or your user account.

Creating this resource installs a particular app on a particular repository.

The app installation and the repository must both belong to the same
organization or user account on GitHub. Note: you can review your organization's installations
by the following the instructions at this
[link](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/reviewing-your-organizations-installed-integrations) or for your user account at this [link](https://docs.github.com/en/apps/using-github-apps/reviewing-and-modifying-installed-github-apps).

## Example Usage

```hcl
# Create a repository.
resource "github_repository" "some_repo" {
  name = "some-repo"
}

resource "github_app_installation_repository" "some_app_repo" {
  # The installation id of the app (in the organization).
  installation_id    = "1234567"
  repository         = "${github_repository.some_repo.name}"
}
```

## Argument Reference

The following arguments are supported:

* `installation_id` - (Required) The GitHub app installation id.
* `repository`      - (Required) The repository to install the app on.

## Import

GitHub App Installation Repository can be imported
using an ID made up of `installation_id:repository`, e.g.

```
$ terraform import github_app_installation_repository.terraform_repo 1234567:terraform
```
