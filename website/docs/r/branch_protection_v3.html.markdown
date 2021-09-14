---
layout: "github"
page_title: "GitHub:  github_branch_protection_v3"
description: |-
  Protects a GitHub branch using the v3 / REST implementation.  The `github_branch_protection` resource has moved to the GraphQL API, while this resource will continue to leverage the REST API
---

# github\_branch\_protection\_v3

Protects a GitHub branch.

The `github_branch_protection` resource has moved to the GraphQL API, while this resource will continue to leverage the REST API.

This resource allows you to configure branch protection for repositories in your organization. When applied, the branch will be protected from forced pushes and deletion. Additional constraints, such as required status checks or restrictions on users, teams, and apps, can also be configured.

## Example Usage

```hcl
# Protect the main branch of the foo repository. Only allow a specific user to merge to the branch.
resource "github_branch_protection_v3" "example" {
  repository     = github_repository.example.name
  branch         = "main"

  restrictions {
    users = ["foo-user"]
  }
}
```

```hcl
# Protect the main branch of the foo repository. Additionally, require that
# the "ci/travis" context to be passing and only allow the engineers team merge
# to the branch.

resource "github_branch_protection_v3" "example" {
  repository     = github_repository.example.name
  branch         = "main"
  enforce_admins = true

  required_status_checks {
    strict   = false
    contexts = ["ci/travis"]
  }

  required_pull_request_reviews {
    dismiss_stale_reviews = true
    dismissal_users       = ["foo-user"]
    dismissal_teams       = [github_team.example.slug]
  }

  restrictions {
    users = ["foo-user"]
    teams = [github_team.example.slug]
    apps  = ["foo-app"]
  }
}

resource "github_repository" "example" {
  name = "example"
}

resource "github_team" "example" {
  name = "Example Name"
}

resource "github_team_repository" "example" {
  team_id    = github_team.example.id
  repository = github_repository.example.name
  permission = "pull"
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) The GitHub repository name.
* `branch` - (Required) The Git branch to protect.
* `enforce_admins` - (Optional) Boolean, setting this to `true` enforces status checks for repository administrators.
* `require_signed_commits` - (Optional) Boolean, setting this to `true` requires all commits to be signed with GPG.
* `require_conversation_resolution` - (Optional) Boolean, setting this to `true` requires all conversations on code must be resolved before a pull request can be merged.
* `required_status_checks` - (Optional) Enforce restrictions for required status checks. See [Required Status Checks](#required-status-checks) below for details.
* `required_pull_request_reviews` - (Optional) Enforce restrictions for pull request reviews. See [Required Pull Request Reviews](#required-pull-request-reviews) below for details.
* `restrictions` - (Optional) Enforce restrictions for the users and teams that may push to the branch. See [Restrictions](#restrictions) below for details.

### Required Status Checks

`required_status_checks` supports the following arguments:

* `strict`: (Optional) Require branches to be up to date before merging. Defaults to `false`.
* `contexts`: (Optional) The list of status checks to require in order to merge into this branch. No status checks are required by default.

### Required Pull Request Reviews

`required_pull_request_reviews` supports the following arguments:

* `dismiss_stale_reviews`: (Optional) Dismiss approved reviews automatically when a new commit is pushed. Defaults to `false`.
* `dismissal_users`: (Optional) The list of user logins with dismissal access
* `dismissal_teams`: (Optional) The list of team slugs with dismissal access.
  Always use `slug` of the team, **not** its name. Each team already **has** to have access to the repository.
* `require_code_owner_reviews`: (Optional) Require an approved review in pull requests including files with a designated code owner. Defaults to `false`.
* `required_approving_review_count`: (Optional) Require x number of approvals to satisfy branch protection requirements. If this is specified it must be a number between 1-6. This requirement matches GitHub's API, see the upstream [documentation](https://developer.github.com/v3/repos/branches/#parameters-1) for more information.

### Restrictions

`restrictions` supports the following arguments:

* `users`: (Optional) The list of user logins with push access.
* `teams`: (Optional) The list of team slugs with push access.
  Always use `slug` of the team, **not** its name. Each team already **has** to have access to the repository.
* `apps`: (Optional) The list of app slugs with push access.

`restrictions` is only available for organization-owned repositories.

## Import

GitHub Branch Protection can be imported using an ID made up of `repository:branch`, e.g.

```
$ terraform import  github_branch_protection_v3.terraform terraform:main
```
