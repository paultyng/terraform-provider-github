---
layout: "github"
page_title: "GitHub: github_membership"
description: |-
  Provides a GitHub membership resource.
---

# github_membership

Provides a GitHub membership resource.

This resource allows you to add/remove users from your organization. When applied,
an invitation will be sent to the user to become part of the organization. When
destroyed, either the invitation will be cancelled or the user will be removed.

~> **NOTE:** This resource **does not** work with GitHub Enterprise which requires
users to be provisioned through an identity provider (IdP).

## Example Usage

```hcl
# Add a user to the organization
resource "github_membership" "membership_for_some_user" {
  username = "SomeUser"
  role     = "member"
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) The user to add to the organization.
* `role` - (Optional) The role of the user within the organization.
            Must be one of `member` or `admin`. Defaults to `member`.
            `admin` role represents the `owner` role available via GitHub UI.
* `downgrade_on_destroy` - (Optional) Defaults to `false`. If set to true,
            when this resource is destroyed, the member will not be removed
            from the organization. Instead, the member's role will be
            downgraded to 'member'.


## Import

GitHub Membership can be imported using an ID made up of `organization:username`, e.g.

```
$ terraform import github_membership.member hashicorp:someuser
```
