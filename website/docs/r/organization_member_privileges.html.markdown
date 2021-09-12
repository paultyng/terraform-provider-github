---
layout: "github"
page_title: "GitHub: github_organization_member_privileges"
description: |-
  Manages member privileges in GitHub organizations
---

# github_organization_member_privileges

This resource allows you to manage member privileges in a GitHub organization.

## Example Usage

```hcl
resource "github_organization_member_privileges" "organization" {
  default_repository_permission   = "none"
  members_can_create_repositories = false
}
```

## Argument Reference

The following arguments are supported:

* `default_repository_permission` - (Optional) Default permission level members have for organization repositories. Can be `read` (can pull, but not push to or administer this repository), `write` (can pull and push, but not administer this repository), `admin` (can pull, push, and administer this repository) or `none` (no permissions granted by default).

* `members_can_create_repositories` - (Optional) Toggles the ability of non-admin organization members to create repositories.

* `members_can_create_internal_repositories` - (Optional) Toggles whether organization members can create internal repositories, which are visible to all enterprise members. You can only allow members to create internal repositories if your organization is associated with an enterprise account using GitHub Enterprise Cloud or GitHub Enterprise Server 2.20+. We detect if you plan is compatible with this option by looking at the name of you plan, if the name of you plan is `free` this option will not be applied.

* `members_can_create_private_repositories` - (Optional) Toggles whether organization members can create private repositories, which are visible to organization members with permission.

* `members_can_create_public_repositories` - (Optional) Toggles whether organization members can create public repositories, which are visible to anyone.


## Attributes Reference

The following additional attributes are exported:

* `etag` - An etag representing the Organization object.
