---
layout: "github"
page_title: "github_organization_ruleset Resource - terraform-provider-github"
description: |-
  Creates a GitHub organization ruleset.
---

# github_organization_ruleset (Resource)

Creates a GitHub organization ruleset.

This resource allows you to create and manage rulesets on the organization level. When applied, a new ruleset will be created. When destroyed, that ruleset will be removed.

## Example Usage

```hcl
resource "github_organization_ruleset" "example" {
  name        = "example"
  target      = "branch"
  enforcement = "active"

  conditions {
    ref_name {
      include = ["~ALL"]
      exclude = []
    }
  }

  bypass_actors {
    actor_id    = 13473
    actor_type  = "Integration"
    bypass_mode = "always"
  }

  rules {
    creation                = true
    update                  = true
    deletion                = true
    required_linear_history = true
    required_signatures     = true

    branch_name_pattern {
      name     = "example"
      negate   = false
      operator = "starts_with"
      pattern  = "ex"
    }
  }
}
```

## Argument Reference

* `enforcement` - (Required) (String) Possible values for Enforcement are `disabled`, `active`, `evaluate`. Note: `evaluate` is currently only supported for owners of type `organization`.

* `name` - (Required) (String) The name of the ruleset.

* `rules` - (Required) (Block List, Min: 1, Max: 1) Rules within the ruleset. (see [below for nested schema](#rules))

* `target` - (Required) (String) Possible values are `branch` and `tag`.

* `bypass_actors` - (Optional) (Block List) The actors that can bypass the rules in this ruleset. (see [below for nested schema](#bypass_actors))

* `conditions` - (Optional) (Block List, Max: 1) Parameters for an organization ruleset condition. `ref_name` is required alongside one of `repository_name` or `repository_id`. (see [below for nested schema](#conditions))

#### Rules ####

The `rules` block supports the following:

* `branch_name_pattern` - (Optional) (Block List, Max: 1) Parameters to be used for the branch_name_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations. Conflicts with `tag_name_pattern` as it only applies to rulesets with target `branch`. (see [below for nested schema](#rules.branch_name_pattern))

* `commit_author_email_pattern` - (Optional) (Block List, Max: 1) Parameters to be used for the commit_author_email_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations. (see [below for nested schema](#rules.commit_author_email_pattern))

* `commit_message_pattern` - (Optional) (Block List, Max: 1) Parameters to be used for the commit_message_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations. (see [below for nested schema](#rules.commit_message_pattern))

* `committer_email_pattern` - (Optional) (Block List, Max: 1) Parameters to be used for the committer_email_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations. (see [below for nested schema](#rules.committer_email_pattern))

* `creation` - (Optional) (Boolean) Only allow users with bypass permission to create matching refs.

* `deletion` - (Optional) (Boolean) Only allow users with bypass permissions to delete matching refs.

* `non_fast_forward` - (Optional) (Boolean) Prevent users with push access from force pushing to branches.

* `pull_request` - (Optional) (Block List, Max: 1) Require all commits be made to a non-target branch and submitted via a pull request before they can be merged. (see [below for nested schema](#rules.pull_request))

* `required_linear_history` - (Optional) (Boolean) Prevent merge commits from being pushed to matching branches.

* `required_signatures` - (Optional) (Boolean) Commits pushed to matching branches must have verified signatures.

* `required_status_checks` - (Optional) (Block List, Max: 1) Choose which status checks must pass before branches can be merged into a branch that matches this rule. When enabled, commits must first be pushed to another branch, then merged or pushed directly to a branch that matches this rule after status checks have passed. (see [below for nested schema](#rules.required_status_checks))

* `required_workflows` - (Optional) (Block List, Max: 1) Define which Actions workflows must pass before changes can be merged into a branch matching the rule. Multiple workflows can be specified. (see [below for nested schema](#rules.required_workflows))

* `required_code_scanning` - (Optional) (Block List, Max: 1) Define which tools must provide code scanning results before the reference is updated. When configured, code scanning must be enabled and have results for both the commit and the reference being updated. Multiple code scanning tools can be specified. (see [below for nested schema](#rules.required_code_scanning))

* `tag_name_pattern` - (Optional) (Block List, Max: 1) Parameters to be used for the tag_name_pattern rule. This rule only applies to repositories within an enterprise, it cannot be applied to repositories owned by individuals or regular organizations. Conflicts with `branch_name_pattern` as it only applies to rulesets with target `tag`. (see [below for nested schema](#rules.tag_name_pattern))

* `update` - (Optional) (Boolean) Only allow users with bypass permission to update matching refs.

#### rules.branch_name_pattern ####

* `operator` - (Required) (String) The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.

* `pattern` - (Required) (String) The pattern to match with.

* `name` - (Optional) (String) How this rule will appear to users.

* `negate` - (Optional) (Boolean) If true, the rule will fail if the pattern matches.

#### rules.commit_author_email_pattern ####

* `operator` - (Required) (String) The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.

* `pattern` - (Required) (String) The pattern to match with.

* `name` - (Optional) (String) How this rule will appear to users.

* `negate` - (Optional) (Boolean) If true, the rule will fail if the pattern matches.

#### rules.commit_message_pattern ####

* `operator` - (Required) (String) The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.

* `pattern` - (Required) (String) The pattern to match with.

* `name` - (Optional) (String) How this rule will appear to users.

* `negate` - (Optional) (Boolean) If true, the rule will fail if the pattern matches.

#### rules.committer_email_pattern ####

* `operator` - (Required) (String) The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.

* `pattern` - (Required) (String) The pattern to match with.

* `name` - (Optional) (String) How this rule will appear to users.

* `negate` - (Optional) (Boolean) If true, the rule will fail if the pattern matches.

#### rules.pull_request ####

* `dismiss_stale_reviews_on_push` - (Optional) (Boolean) New, reviewable commits pushed will dismiss previous pull request review approvals. Defaults to `false`.

* `require_code_owner_review` - (Optional) (Boolean) Require an approving review in pull requests that modify files that have a designated code owner. Defaults to `false`.

* `require_last_push_approval` - (Optional) (Boolean) Whether the most recent reviewable push must be approved by someone other than the person who pushed it. Defaults to `false`.

* `required_approving_review_count` - (Optional) (Number) The number of approving reviews that are required before a pull request can be merged. Defaults to `0`.

* `required_review_thread_resolution` - (Optional) (Boolean) All conversations on code must be resolved before a pull request can be merged. Defaults to `false`.

#### rules.required_status_checks ####

* `required_check` - (Required) (Block Set, Min: 1) Status checks that are required. Several can be defined. (see [below for nested schema](#rules.required_status_checks.required_check))

* `strict_required_status_checks_policy` - (Optional) (Boolean) Whether pull requests targeting a matching branch must be tested with the latest code. This setting will not take effect unless at least one status check is enabled. Defaults to `false`.

#### required_status_checks.required_check ####

* `context` - (Required) (String) The status check context name that must be present on the commit.

* `integration_id` - (Optional) (Number) The optional integration ID that this status check must originate from.

#### rules.required_workflows ####

* `required_workflow` - (Required) (Block Set, Min: 1) Actions workflows that are required. Multiple can be defined. (see [below for nested schema](#rules.required_workflows.required_workflow))

#### rules.required_workflows.required_workflow ####

* `repository_id` - (Required) (Number) The ID of the repository. Names, full names and repository URLs are not supported.

* `path` - (Required) (String) The path to the YAML definition file of the workflow.

* `ref` - (Optional) (String) The optional ref from which to fetch the workflow. Defaults to `master`.

#### rules.required_code_scanning ####

* `required_code_scanning_tool` - (Required) (Block Set, Min: 1) Actions code scanning tools that are required. Multiple can be defined. (see [below for nested schema](#rules.required_workflows.required_code_scanning_tool))

#### rules.required_code_scanning.required_code_scanning_tool ####

* `alerts_threshold` - (Required) (String) The severity level at which code scanning results that raise alerts block a reference update. Can be one of: `none`, `errors`, `errors_and_warnings`, `all`.

* `security_alerts_threshold` - (Required) (String) The severity level at which code scanning results that raise security alerts block a reference update. Can be one of: `none`, `critical`, `high_or_higher`, `medium_or_higher`, `all`.

* `tool` - (Required) (String) The name of a code scanning tool.

#### rules.tag_name_pattern ####

* `operator` - (Required) (String) The operator to use for matching. Can be one of: `starts_with`, `ends_with`, `contains`, `regex`.

* `pattern` - (Required) (String) The pattern to match with.

* `name` - (Optional) (String) How this rule will appear to users.

* `negate` - (Optional) (Boolean) If true, the rule will fail if the pattern matches.

#### bypass_actors ####

* `actor_id` - (Required) (Number) The ID of the actor that can bypass a ruleset.

* `actor_type` (String) The type of actor that can bypass a ruleset. Can be one of: `RepositoryRole`, `Team`, `Integration`, `OrganizationAdmin`.

* `bypass_mode` - (Optional) (String) When the specified actor can bypass the ruleset. pull_request means that an actor can only bypass rules on pull requests. Can be one of: `always`, `pull_request`.

~>Note: at the time of writing this, the following actor types correspond to the following actor IDs:

* `OrganizationAdmin` -> `1`
* `RepositoryRole` (This is the actor type, the following are the base repository roles and their associated IDs.)
  * `maintain` -> `2`
  * `write` -> `4`
  * `admin` -> `5`

#### conditions ####

* `ref_name` - (Required) (Block List, Min: 1, Max: 1) (see [below for nested schema](#conditions.ref_name))
* `repository_id` (Optional) (List of Number) The repository IDs that the ruleset applies to. One of these IDs must match for the condition to pass. Conflicts with `repository_name` and `repository_property`.
* `repository_name` (Optional) (Block List, Max: 1) Conflicts with `repository_id` and `repository_property`. (see [below for nested schema](#conditions.repository_name))
* `repository_property` (Optional) (Block List, Max: 1) Conflicts with `repository_id` and `repository_name`. (see [below for nested schema](#conditions.repository_property))

One of `repository_id` and `repository_name` must be set for the rule to target any repositories.

#### conditions.ref_name ####

* `exclude` - (Required) (List of String) Array of ref names or patterns to exclude. The condition will not pass if any of these patterns match.

* `include` - (Required) (List of String) Array of ref names or patterns to include. One of these patterns must match for the condition to pass. Also accepts `~DEFAULT_BRANCH` to include the default branch or `~ALL` to include all branches.

#### conditions.repository_name ####

* `exclude` - (Required) (List of String) Array of repository names or patterns to exclude. The condition will not pass if any of these patterns match.

* `include` - (Required) (List of String) Array of repository names or patterns to include. One of these patterns must match for the condition to pass. Also accepts `~ALL` to include all repositories.

#### conditions.repository_property ####

* `exclude` - (Optional) (List of Repository Properties) The repository properties and values to include. All of these properties must match for the condition to pass. (see [below for nested schema](#conditions.repository_property.properties))

* `include` - (Optional) (List of Repository Properties) The repository properties and values to exclude. The condition will not pass if any of these properties match.(see [below for nested schema](#conditions.repository_property.properties))

#### conditions.repository_property.properties ####
* `name` (Required) (String) The name of the repository property to target.

* `property_values` (Required) (Array of String) The values to match for the repository property.

* `source` (String) The source of the repository property. Defaults to 'custom' if not specified. Can be one of: `custom`, `system`

## Attributes Reference

The following additional attributes are exported:

* `etag` (String)

* `node_id` (String) GraphQL global node id for use with v4 API.

* `ruleset_id` (Number) GitHub ID for the ruleset.

## Import

GitHub Organization Rulesets can be imported using the GitHub ruleset ID e.g.

`$ terraform import github_organization_ruleset.example 12345`
