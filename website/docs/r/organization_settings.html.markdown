---
layout: "github"
page_title: "GitHub: github_organization_settings"
description: |-
  Creates and manages settings for a GitHub Organization.
---

# github_organization_settings

This resource allows you to create and manage settings for a GitHub Organization.

## Example Usage

```hcl
resource "github_organization_settings" "test" {
    billing_email = "test@example.com"
    company = "Test Company"
    blog = "https://example.com"
    email = "test@example.com"
    twitter_username = "Test"
    location = "Test Location"
    name = "Test Name"
    description = "Test Description"
    has_organization_projects = true
    has_repository_projects = true
    default_repository_permission = "read"
    members_can_create_repositories = true
    members_can_create_public_repositories = true
    members_can_create_private_repositories = true
    members_can_create_internal_repositories = true
    members_can_create_pages = true
    members_can_create_public_pages = true
    members_can_create_private_pages = true
    members_can_fork_private_repositories = true
    web_commit_signoff_required = true
    advanced_security_enabled_for_new_repositories = false
    dependabot_alerts_enabled_for_new_repositories=  false
    dependabot_security_updates_enabled_for_new_repositories = false
    dependency_graph_enabled_for_new_repositories = false
    secret_scanning_enabled_for_new_repositories = true
    secret_scanning_push_protection_enabled_for_new_repositories = true
    secret_scanning_push_protection_custom_link = "https://github.com/test-org/test-repo/blob/main/README.md"
}
```

## Argument Reference

The following arguments are supported:

* `billing_email` - (Required) The billing email address for the organization.
* `company` - (Optional) The company name for the organization.
* `blog` - (Optional) The blog URL for the organization.
* `email` - (Optional) The email address for the organization.
* `twitter_username` - (Optional) The Twitter username for the organization.
* `location` - (Optional) The location for the organization.
* `name` - (Optional) The name for the organization.
* `description` - (Optional) The description for the organization.
* `has_organization_projects` - (Optional) Whether or not organization projects are enabled for the organization.
* `has_repository_projects` - (Optional) Whether or not repository projects are enabled for the organization.
* `default_repository_permission` - (Optional) The default permission for organization members to create new repositories. Can be one of `read`, `write`, `admin`, or `none`. Defaults to `read`.
* `members_can_create_repositories` - (Optional) Whether or not organization members can create new repositories. Defaults to `true`.
* `members_can_create_public_repositories` - (Optional) Whether or not organization members can create new public repositories. Defaults to `true`.
* `members_can_create_private_repositories` - (Optional) Whether or not organization members can create new private repositories. Defaults to `true`.
* `members_can_create_internal_repositories` - (Optional) Whether or not organization members can create new internal repositories. For Enterprise Organizations only.
* `members_can_create_pages` - (Optional) Whether or not organization members can create new pages. Defaults to `true`.
* `members_can_create_public_pages` - (Optional) Whether or not organization members can create new public pages. Defaults to `true`.
* `members_can_create_private_pages` - (Optional) Whether or not organization members can create new private pages. Defaults to `true`.
* `members_can_fork_private_repositories` - (Optional) Whether or not organization members can fork private repositories. Defaults to `false`.
* `web_commit_signoff_required` - (Optional) Whether or not commit signatures are required for commits to the organization. Defaults to `false`.
* `advanced_security_enabled_for_new_repositories` - (Optional) Whether or not advanced security is enabled for new repositories. Defaults to `false`.
* `dependabot_alerts_enabled_for_new_repositories` - (Optional) Whether or not dependabot alerts are enabled for new repositories. Defaults to `false`.
* `dependabot_security_updates_enabled_for_new_repositories` - (Optional) Whether or not dependabot security updates are enabled for new repositories. Defaults to `false`.
* `dependency_graph_enabled_for_new_repositories` - (Optional) Whether or not dependency graph is enabled for new repositories. Defaults to `false`.
* `secret_scanning_enabled_for_new_repositories` - (Optional) Whether or not secret scanning is enabled for new repositories. Defaults to `false`.
* `secret_scanning_push_protection_enabled_for_new_repositories` - (Optional) Whether or not secret scanning push protection is enabled for new repositories. Defaults to `false`.
* `secret_scanning_push_protection_custom_link` - (Optional) The URL that will be displayed to contributors who are blocked from pushing a secret. Defaults to `null`.


## Attributes Reference

The following additional attributes are exported:

* `id` - The ID of the organization settings.


## Import

Organization settings can be imported using the `id` of the organization.
The `id` of the organization can be found using the [get an organization](https://docs.github.com/en/rest/orgs/orgs#get-an-organization) API.

```
$ terraform import github_organization_settings.test 123456789
```
