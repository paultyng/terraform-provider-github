---
layout: "github"
page_title: "GitHub: github_repository_custom_properties"
description: |-
  Get all custom properties of a repository
---

# github_repository_custom_properties

Use this data source to retrieve all custom properties of a repository.

## Example Usage

```hcl
data "github_repository_custom_properties" "example" {
    repository = "example-repository"
}
```

## Argument Reference

* `repository` - (Required) Name of the repository to retrieve the custom properties from.

## Attributes Reference

* `property` - The list of this repository's custom properties. Each element of `property` has the following attributes:
    * `property_name` - Name of the property
    * `property_value` - Value of the property
