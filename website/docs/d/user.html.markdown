---
layout: "github"
page_title: "GitHub: github_user"
description: |-
  Get information on a GitHub user.
---

# github\_user

Use this data source to retrieve information about a GitHub user.

## Example Usage

```hcl
# Retrieve information about a GitHub user.
data "github_user" "example" {
  username = "example"
}

# Retrieve information about a GitHub user by ID.
data "github_user" "example" {
  id = 1
}

# Retrieve information about the currently authenticated user.
data "github_user" "current" {
  username = ""
}

output "current_github_login" {
  value = "${data.github_user.current.login}"
}

```

## Argument Reference

 * `username` - (Optional) The username. Use an empty string `""` to retrieve information about the currently authenticated user.
 * `id` - (Optional) The id.

 -> Exactly one of `username` or `id` must be specified.

## Attributes Reference

 * `node_id` - the Node ID of the user.
 * `login` - the user's login.
 * `avatar_url` - the user's avatar URL.
 * `gravatar_id` - the user's gravatar ID.
 * `site_admin` - whether the user is a GitHub admin.
 * `name` - the user's full name.
 * `company` - the user's company name.
 * `blog` - the user's blog location.
 * `location` - the user's location.
 * `email` - the user's email.
 * `gpg_keys` - list of user's GPG keys.
 * `ssh_keys` - list of user's SSH keys.
 * `bio` - the user's bio.
 * `public_repos` - the number of public repositories.
 * `public_gists` - the number of public gists.
 * `followers` - the number of followers.
 * `following` - the number of following users.
 * `created_at` - the creation date.
 * `updated_at` - the update date.
 * `suspended_at` - the suspended date if the user is suspended.
