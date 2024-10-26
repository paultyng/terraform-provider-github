---
layout: "github"
page_title: "GitHub: github_collaborators"
description: |-
  Get the collaborators for a given repository.
---

# github_collaborators

Use this data source to retrieve the collaborators for a given repository.

## Example Usage

```hcl
data "github_collaborators" "test" {
  owner      = "example_owner"
  repository = "example_repository"
}
```

## Arguments Reference

 * `owner` - (Required) The organization that owns the repository.

 * `repository` - (Required) The name of the repository.

 * `affiliation` - (Optional) Filter collaborators returned by their affiliation. Can be one of: `outside`, `direct`, `all`.  Defaults to `all`.

 * `permission` - (Optional) Filter collaborators returned by their permission. Can be one of: `pull`, `triage`, `push`, `maintain`, `admin`.  Defaults to not doing any filtering on permission.

## Attributes Reference

 * `collaborator` - An Array of GitHub collaborators.  Each `collaborator` block consists of the fields documented below.

___

The `collaborator` block consists of:

* `login` - The collaborator's login.

* `id` - The ID of the collaborator.

* `url` - The GitHub API URL for the collaborator.

* `html_url` - The GitHub HTML URL for the collaborator.

* `followers_url` - The GitHub API URL for the collaborator's followers.

* `following_url` - The GitHub API URL for those following the collaborator.

* `gists_url` - The GitHub API URL for the collaborator's gists.

* `starred_url` - The GitHub API URL for the collaborator's starred repositories.

* `subscriptions_url` - The GitHub API URL for the collaborator's subscribed repositories.

* `organizations_url` - The GitHub API URL for the collaborator's organizations.

* `repos_url` - The GitHub API URL for the collaborator's repositories.

* `events_url` - The GitHub API URL for the collaborator's events.

* `received_events_url` - The GitHub API URL for the collaborator's received events.

* `type` - The type of the collaborator (ex. `user`).

* `site_admin` - Whether the user is a GitHub admin.

* `permission` - The permission of the collaborator.
