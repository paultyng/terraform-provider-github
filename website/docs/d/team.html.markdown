---
layout: "github"
page_title: "GitHub: github_team"
description: |-
  Get information on a GitHub team.
---

# github\_team

Use this data source to retrieve information about a GitHub team.

## Example Usage

```hcl
data "github_team" "example" {
  slug = "example"
}
```

## Argument Reference

* `slug` - (Required) The team slug.
* `membership_type` - (Optional) Type of membership to be requested to fill the list of members. Can be either "all" or "immediate". Default: "all"
* `summary_only` - (Optional) Exclude the members and repositories of the team from the returned result. Defaults to `false`.
* `results_per_page` - (Optional) Set the number of results per graphql query. Reducing this number can alleviate timeout errors. Accepts a value between 0 - 100. Defaults to `100`.

## Attributes Reference

* `id` - the ID of the team.
* `node_id` - the Node ID of the team.
* `name` - the team's full name.
* `description` - the team's description.
* `privacy` - the team's privacy type.
* `permission` - the team's permission level.
* `members` - List of team members (list of GitHub usernames). Not returned if `summary_only = true`
* `repositories` - (**DEPRECATED**) List of team repositories (list of repo names). Not returned if `summary_only = true`
* `repositories_detailed` - List of team repositories (each item comprises of `repo_id`, `repo_name` & [`role_name`](https://registry.terraform.io/providers/integrations/github/latest/docs/resources/team_repository#permission)). Not returned if `summary_only = true`
