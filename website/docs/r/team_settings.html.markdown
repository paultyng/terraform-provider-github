---
layout: "github"
page_title: "GitHub: github_team_settings"
description: |-
  Manages the team settings (in particular the request review delegation settings)
---

# github_team_settings

This resource manages the team settings (in particular the request review delegation settings) within the organization

Creating this resource will alter the team Code Review settings.

The team must both belong to the same organization configured in the provider on GitHub. 

~> **Note**: This resource relies on the v4 GraphQl GitHub API. If this API is not available, or the Stone Crop schema preview is not available, then this resource will not work as intended.

## Example Usage

```hcl
# Add a repository to the team
resource "github_team" "some_team" {
  name        = "SomeTeam"
  description = "Some cool team"
}

resource "github_team_settings" "code_review_settings" {
  team_id    = github_team.some_team.id
  review_request_delegation {
      algorithm = "ROUND_ROBIN"
      member_count = 1
      notify = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `team_id` - (Required) The GitHub team id or the GitHub team slug
* `review_request_delegation` - (Optional) The settings for delegating code reviews to individuals on behalf of the team. If this block is present, even without any fields, then review request delegation will be enabled for the team. See [GitHub Review Request Delegation](#github-review-request-delegation-configuration) below for details. See [GitHub's documentation](https://docs.github.com/en/organizations/organizing-members-into-teams/managing-code-review-settings-for-your-team#configuring-team-notifications) for more configuration details.

### GitHub Review Request Delegation Configuration

The following arguments are supported:

* `algorithm` - (Optional) The algorithm to use when assigning pull requests to team members. Supported values are `ROUND_ROBIN` and `LOAD_BALANCE`. Default value is `ROUND_ROBIN`
* `member_count` - (Optional) The number of team members to assign to a pull request
* `notify` - (Optional) whether to notify the entire team when at least one member is also assigned to the pull request
* `remove_team_request` - (Optional) whether to remove the team request when a member is assigned to the pull request.


## Import

GitHub Teams can be imported using the GitHub team ID, or the team slug e.g.

```
$ terraform import github_team.code_review_settings 1234567
```
or,
```
$ terraform import github_team_settings.code_review_settings SomeTeam
```
