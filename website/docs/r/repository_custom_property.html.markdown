---
layout: "github"
page_title: "GitHub: github_repository_custom_property"
description: |-
  Creates and a specific custom property for a GitHub repository
---

# github_repository_custom_property

This resource allows you to create and manage a specific custom property for a GitHub repository.

## Example Usage

> Note that this assumes there already is a custom property defined on the org level called `my-cool-property` of type `string`

```hcl
resource "github_repository" "example" {
  name        = "example"
  description = "My awesome codebase"
}
resource "github_repository_custom_property" "string" {
  repository     = github_repository.example.name
  property_name  = "my-cool-property"
  property_type  = "string"
  property_value = ["test"]
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) The repository of the environment.

* `property_type` - (Required) Type of the custom property. Can be one of `single_select`, `multi_select`, `string`, or `true_false`

* `property_name` - (Required) Name of the custom property. Note that a pre-requisiste for this resource is that a custom property of this name has already been defined on the organization level

* `property_value` - (Required) Value of the custom property in the form of an array. Properties of type `single_select`, `string`, and `true_false` are represented as a string array of length 1 

## Import

GitHub Repository Custom Property can be imported using an ID made up of a comibnation of the names of the organization, repository, custom property separated by a `:` character, e.g.

```
$ terraform import github_repository_custom_property.example <organization-name>:<repo-name>:<custom-property-name>
```
