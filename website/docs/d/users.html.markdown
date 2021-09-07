---
layout: "github"
page_title: "GitHub: github_users"
description: |-
  Get information about multiple GitHub users.
---

# github\_users

Use this data source to retrieve information about multiple GitHub users at once.

## Example Usage

```hcl
# Retrieve information about multiple GitHub users.
data "github_users" "example" {
  usernames = ["example1", "example2", "example3"]
}

output "valid_users" {
  value = "${data.github_user.example.logins}"
}

output "invalid_users" {
  value = "${data.github_user.example.unknown_logins}"
}
```

## Argument Reference

 * `usernames` - (Required) List of usernames.

## Attributes Reference

 * `node_ids` - list of Node IDs of users that could be found.
 * `logins` - list of logins of users that could be found.
 * `unknown_logins` - list of logins without matching user.
