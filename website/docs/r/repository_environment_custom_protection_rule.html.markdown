---
layout: "github"
page_title: "GitHub: github_repository_environment_custom_protection_rule"
description: |-
  Creates and manages environment deployment custom protection rules for GitHub repositories
---

# github_repository_environment_custom_protection_rule

This resource allows you to create and manage environment deployment custom protection rules for a GitHub repository.

## Example Usage

```hcl

resource "github_repository" "test" {
  name      = "tf-acc-test-%s"
}

resource "github_repository_environment" "test" {
  repository 	= github_repository.test.name
  environment	= "environment/test"
}

resource "github_repository_environment_custom_protection_rule" "test" {
  repository 	   = github_repository.test.name
  environment	   = github_repository_environment.test.environment
  integration_id = 123456
}
```

## Argument Reference

The following arguments are supported:

* `environment` - (Required) The name of the environment.

* `repository` - (Required) The repository of the environment.

* `integration_id` - (Required) The ID of the custom app that will be enabled on the environment.


## Import

GitHub Repository Environment Deployment Policy can be imported using an ID made up of `name` of the repository combined with the `environment` name of the environment with the `Id` of the protection rule, separated by a `:` character, e.g.

```
$ terraform import github_repository_environment_deployment_policy.daily terraform:daily:123456
```
