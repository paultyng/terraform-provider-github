---
layout: "github"
page_title: "GitHub: github_branch_protection"
description: |-
  Protects a GitHub branch.
---

# github\_branch\_protection

Protects a GitHub branch.

This resource allows you to configure branch protection for repositories in your organization. When applied, the branch will be protected from forced pushes and deletion. Additional constraints, such as required status checks or restrictions on users, teams, and apps, can also be configured.

## Example Usage

```hcl
# Protect the master branch of the foo repository. Additionally, require that
# the "ci/travis" context to be passing and only allow the engineers team merge
# to the branch.

resource "github_branch_protection" "example" {
  repository_id = github_repository.example.node_id
  # also accepts repository name
  # repository_id  = github_repository.example.name

  pattern          = "main"
  enforce_admins   = true
  allows_deletions = true

  required_status_checks {
    strict   = false
    contexts = ["ci/travis"]
  }

  required_pull_request_reviews {
    dismiss_stale_reviews  = true
    restrict_dismissals    = true
    dismissal_restrictions = [
      data.github_user.example.node_id,
      github_team.example.node_id,
    ]
  }

  push_restrictions = [
    data.github_user.example.node_id,
    # limited to a list of one type of restriction (user, team, app)
    # github_team.example.node_id
  ]

}

resource "github_repository" "example" {
  name = "test"
}

data "github_user" "example" {
  username = "example"
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

* `repository_id` - (Required) The name or node ID of the repository associated with this branch protection rule.
* `pattern` - (Required) Identifies the protection rule pattern.
* `enforce_admins` - (Optional) Boolean, setting this to `true` enforces status checks for repository administrators.
* `require_signed_commits` - (Optional) Boolean, setting this to `true` requires all commits to be signed with GPG.
* `require_conversation_resolution` - (Optional) Boolean, setting this to `true` requires all conversations on code must be resolved before a pull request can be merged.
* `required_status_checks` - (Optional) Enforce restrictions for required status checks. See [Required Status Checks](#required-status-checks) below for details.
* `required_pull_request_reviews` - (Optional) Enforce restrictions for pull request reviews. See [Required Pull Request Reviews](#required-pull-request-reviews) below for details.
* `push_restrictions` - (Optional) The list of actor IDs that may push to the branch.
* `allows_deletions` - (Optional) Boolean, setting this to `true` to allow the branch to be deleted.
* `allows_force_pushes` - (Optional) Boolean, setting this to `true` to allow force pushes on the branch.

### Required Status Checks

`required_status_checks` supports the following arguments:

* `strict`: (Optional) Require branches to be up to date before merging. Defaults to `false`.
* `contexts`: (Optional) The list of status checks to require in order to merge into this branch. No status checks are required by default.

### Required Pull Request Reviews

`required_pull_request_reviews` supports the following arguments:

* `dismiss_stale_reviews`: (Optional) Dismiss approved reviews automatically when a new commit is pushed. Defaults to `false`.
* `restrict_dismissals`: (Optional) Restrict pull request review dismissals.
* `dismissal_restrictions`: (Optional) The list of actor IDs with dismissal access. If not empty, `restrict_dismissals` is ignored.
* `require_code_owner_reviews`: (Optional) Require an approved review in pull requests including files with a designated code owner. Defaults to `false`.
* `required_approving_review_count`: (Optional) Require x number of approvals to satisfy branch protection requirements. If this is specified it must be a number between 1-6. This requirement matches GitHub's API, see the upstream [documentation](https://developer.github.com/v3/repos/branches/#parameters-1) for more information.


## Import

GitHub Branch Protection can be imported using an ID made up of `repository:pattern`, e.g.

```
$ terraform import github_branch_protection.terraform terraform:main
```
