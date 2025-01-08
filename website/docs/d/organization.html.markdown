---
layout: "github"
page_title: "GitHub: github_organization"
description: |-
  Get an organization.
---

# github_organization

Use this data source to retrieve basic information about a GitHub Organization.

## Example Usage

```hcl
data "github_organization" "example" {
  name = "github"
}
```

## Argument Reference

* `name` - (Required) The name of the organization.
* `ignore_archived_repos` - (Optional) Whether or not to include archived repos in the `repositories` list. Defaults to `false`.
* `summary_only` - (Optional) Exclude the repos, members and other attributes from the returned result. Defaults to `false`.

## Attributes Reference

 * `id` - The ID of the organization
 * `node_id` - GraphQL global node ID for use with the v4 API
 * `name` - The organization's public profile name
 * `orgname` - The organization's name as used in URLs and the API
 * `login` - The organization account login
 * `description` - The organization account description
 * `plan` - The organization account plan name
 * `repositories` - (`list`) A list of the full names of the repositories in the organization formatted as `owner/name` strings
 * `members` - **Deprecated**: use `users` instead by replacing `github_organization.example.members` to `github_organization.example.users[*].login` which will give you the same value, expect this field to be removed in next major version
 * `users` - (`list`) A list with the members of the organization with following fields:
   * `id` - The ID of the member
   * `login` - The members login
   * `email` - Publicly available email
   * `role` - Member role `ADMIN`, `MEMBER`
 * `two_factor_requirement_enabled` - Whether two-factor authentication is required for all members of the organization.
 * `default_repository_permission` - Default permission level members have for organization repositories.
 * `members_allowed_repository_creation_type` - The type of repository allowed to be created by members of the organization. Can be one of `ALL`, `PUBLIC`, `PRIVATE`, `NONE`.
 * `members_can_create_repositories` - Whether non-admin organization members can create repositories.
 * `members_can_create_internal_repositories` - Whether organization members can create internal repositories.
 * `members_can_create_private_repositories` - Whether organization members can create private repositories.
 * `members_can_create_public_repositories` - Whether organization members can create public repositories.
 * `members_can_create_pages` - Whether organization members can create pages sites.
 * `members_can_create_public_pages` - Whether organization members can create public pages sites.
 * `members_can_create_private_pages` - Whether organization members can create private pages sites.
 * `members_can_fork_private_repositories` - Whether organization members can create private repository forks.
 * `web_commit_signoff_required` - Whether organization members must sign all commits.
 * `advanced_security_enabled_for_new_repositories` - Whether advanced security is enabled for new repositories.
 * `dependabot_alerts_enabled_for_new_repositories` - Whether Dependabot alerts is automatically enabled for new repositories.
 * `dependabot_security_updates_enabled_for_new_repositories` - Whether Dependabot security updates is automatically enabled for new repositories.
 * `dependency_graph_enabled_for_new_repositories` - Whether dependency graph is automatically enabled for new repositories.
 * `secret_scanning_enabled_for_new_repositories` - Whether secret scanning is automatically enabled for new repositories.
 * `secret_scanning_push_protection_enabled_for_new_repositories` - Whether secret scanning push protection is automatically enabled for new repositories.
 * `secret_scanning_push_protection_custom_link` - The URL that will be displayed to contributors who are blocked from pushing a secret.
