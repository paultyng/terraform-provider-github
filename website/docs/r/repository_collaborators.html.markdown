---
layout: "github"
page_title: "GitHub: github_repository_collaborators"
description: |-
  Provides a GitHub repository collaborators resource.
---

# github_repository_collaborators

Provides a GitHub repository collaborators resource.

~> Note: github_repository_collaborators cannot be used in conjunction with github_repository_collaborator and
github_team_repository or they will fight over what your policy should be.

This resource allows you to manage all collaborators for repositories in your
organization or personal account. For organization repositories, collaborators can
have explicit (and differing levels of) read, write, or administrator access to
specific repositories, without giving the user full organization membership.
For personal repositories, collaborators can only be granted write
(implicitly includes read) permission.

When applied, an invitation will be sent to the user to become a collaborators
on a repository. When destroyed, either the invitation will be cancelled or the
collaborators will be removed from the repository.

This resource is authoritative. For adding a collaborator to a repo in a non-authoritative manner, use
github_repository_collaborator instead.

Further documentation on GitHub collaborators:

- [Adding outside collaborators to your personal repositories](https://help.github.com/en/github/setting-up-and-managing-your-github-user-account/managing-access-to-your-personal-repositories)
- [Adding outside collaborators to repositories in your organization](https://help.github.com/articles/adding-outside-collaborators-to-repositories-in-your-organization/)
- [Converting an organization member to an outside collaborators](https://help.github.com/articles/converting-an-organization-member-to-an-outside-collaborator/)

## Example Usage

```hcl
# Add collaborators to a repository
resource "github_team" "some_team" {
  name        = "SomeTeam"
  description = "Some cool team"
}

resource "github_repository" "some_repo" {
  name = "some-repo"
}

resource "github_repository_collaborators" "some_repo_collaborators" {
  repository = github_repository.some_repo.name

  user {
    permission = "admin"
    username  = "SomeUser"
  }

  team {
    permission = "pull"
    team_id = github_team.some_team.slug
  }
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) The GitHub repository.
* `user` - (Optional) List of users to grant access to the repository.
* `team` - (Optional) List of teams to grant access to the repository.
* `ignore_team` - (Optional) List of teams to ignore when checking for repository access. This supports ignoring teams granted access at an organizational level.

The `user` block supports:

* `username` - (Required) The user to add to the repository as a collaborator.
* `permission` - (Optional) The permission of the outside collaborators for the repository.
            Must be one of `pull`, `push`, `maintain`, `triage` or `admin` or the name of an existing [custom repository role](https://docs.github.com/en/enterprise-cloud@latest/organizations/managing-peoples-access-to-your-organization-with-roles/managing-custom-repository-roles-for-an-organization) within the organization for organization-owned repositories.
            Must be `push` for personal repositories. Defaults to `push`.

The `team` block supports:

* `team_id` - (Required) The GitHub team id or the GitHub team slug.
* `permission` - (Optional) The permission of the outside collaborators for the repository.
  Must be one of `pull`, `triage`, `push`, `maintain`, `admin` or the name of an existing [custom repository role](https://docs.github.com/en/enterprise-cloud@latest/organizations/managing-peoples-access-to-your-organization-with-roles/managing-custom-repository-roles-for-an-organization) within the organisation. Defaults to `pull`.
  Must be `push` for personal repositories. Defaults to `push`.

The `team_ignore` block supports:

* `team_id` - (Required) The GitHub team id or the GitHub team slug.

## Attribute Reference

In addition to the above arguments, the following attributes are exported:

* `invitation_ids` - Map of usernames to invitation ID for any users added as part of creation of this resource to
  be used in [`github_user_invitation_accepter`](./user_invitation_accepter.html).

## Import

GitHub Repository Collaborators can be imported using the name `name`, e.g.

```
$ terraform import github_repository_collaborators.collaborators terraform
```
