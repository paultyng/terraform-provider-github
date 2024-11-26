---
layout: "github"
page_title: "GitHub: github_actions_environment_public_key"
description: |-
  Get Actions public key of the repository environment
---

# github\_actions\_environment\_public\_key

Use this data source to retrieve information about the public key of the repository environment.
Note that the provider `token` must have the "Secrets" repository permissions (read) and "Environments" repository permissions (read) to retrieve the public key of a given environment.

## Example Usage

```hcl
resource "github_repository" "example" {
    name      = "example"
    auto_init = true
}

resource "github_repository_environment" "example" {
    repository       = github_repository.example.name
    environment      = "example"
}

data "github_actions_environment_public_key" "example" {
    repository_id	= github_repository.test.repo_id
    environment     = github_repository_environment.example.environment
}
```

## Argument Reference

* `repository_id` - (Required) The repository ID
* `environment` - (Required) The repository's environment name

## Attributes Reference

* `key_id` - ID of the key that has been retrieved.
* `key` - Actual key retrieved.
