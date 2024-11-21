---
layout: "github"
page_title: "GitHub: github_app_installation_repository"
description: |-
  Manages the associations between app installations and repositories.
---

# github_app_installation_repositories

~> **Note**: This resource is not compatible with the GitHub App Installation authentication method.

This resource manages relationships between app installations and repositories
in your GitHub organization or your user account.

Creating this resource installs a particular app on multiple repositories.

The app installation and the repositories must all belong to the same
organization or user account on GitHub. Note: you can review your organization's installations
by the following the instructions at this
[link](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/reviewing-your-organizations-installed-integrations) or for your user account at this [link](https://docs.github.com/en/apps/using-github-apps/reviewing-and-modifying-installed-github-apps).

## Example Usage

```hcl
# Create some repositories.
resource "github_repository" "some_repo" {
  name = "some-repo"
}

resource "github_repository" "another_repo" {
  name = "another-repo"
}

resource "github_app_installation_repositories" "some_app_repos" {
  # The installation id of the app (in the organization).
  installation_id        = "1234567"
  selected_repositories  = [github_repository.some_repo.name, github_repository.another_repo.name]"
}
```

## Argument Reference

The following arguments are supported:

* `installation_id`       - (Required) The GitHub app installation id.
* `selected_repositories` - (Required) A list of repository names to install the app on.

~> **Note**: Due to how GitHub implements app installations, apps cannot be installed with no repositories selected. Therefore deleting this resource will leave one repository with the app installed. Manually uninstall the app or set the installation to all repositories via the GUI as after deleting this resource.

## Import

GitHub App Installation Repositories can be imported
using an ID made up of `installation_id`, e.g.

```
$ terraform import github_app_installation_repositories.some_app_repos 1234567
```
