---
layout: "github"
page_title: "GitHub: github_organization_teams"
description: |-
  Get information on all GitHub teams of an organization.
---

# github\_organization\_teams

Use this data source to retrieve information about all GitHub teams in an organization.

## Example Usage

To retrieve *all* teams of the organization:

```hcl
data "github_organization_teams" "all" {}
```

To retrieve only the team's at the root of the organization:

```hcl
data "github_organization_teams" "root_teams" {
  root_teams_only = true
}
```

## Attributes Reference

* `teams` - (Required) An Array of GitHub Teams.  Each `team` block consists of the fields documented below.
* `root_teams_only` - (Optional) Only return teams that are at the organization's root, i.e. no nested teams. Defaults to `false`.
* `summary_only` - (Optional) Exclude the members and repositories of the team from the returned result. Defaults to `false`.
* `results_per_page` - (Optional) Set the number of results per graphql query. Reducing this number can alleviate timeout errors. Accepts a value between 0 - 100. Defaults to `100`.

___

The `team` block consists of:

 * `id` - The ID of the team.
 * `node_id` - The Node ID of the team.
 * `slug` - The slug of the team.
 * `name` - The team's full name.
 * `description` - The team's description.
 * `privacy` - The team's privacy type.
 * `members` - List of team members. Not returned if `summary_only = true`
 * `repositories` - List of team repositories. Not returned if `summary_only = true`
 * `parent_team_id` - The ID of the parent team, if there is one.
 * `parent_team_slug` - The slug of the parent team, if there is one.
 * `parent` - (**DEPRECATED**) The parent team, use `parent_team_id` or `parent_team_slug` instead.
