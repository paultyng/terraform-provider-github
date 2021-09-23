---
layout: "github"
page_title: "GitHub: github_release"
description: |-
  Get information on a GitHub release.
---

# github\_release

Use this data source to retrieve information about a GitHub release in a specific repository.

## Example Usage
To retrieve the latest release that is present in a repository:

```hcl
data "github_release" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    retrieve_by = "latest"
}
```

To retrieve a specific release from a repository based on it's ID:

```hcl
data "github_release" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    retrieve_by = "id"
    id          = 12345
}
```

Finally, to retrieve a release based on it's tag:

```hcl
data "github_release" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    retrieve_by = "tag"
    release_tag = "v1.0.0"
}
```

## Argument Reference

 *  `repository`  -  (Required) Name of the repository to retrieve the release from.

 *  `owner`  -  (Required) Owner of the repository.

 *  `retrieve_by`  -  (Required) Describes how to fetch the release. Valid values are `id`, `tag`, `latest`.

 *  `release_id`  -  (Optional) ID of the release to retrieve. Must be specified when `retrieve_by` = `id`.

 *  `release_tag`  -  (Optional) Tag of the release to retrieve. Must be specified when `retrieve_by` = `tag`.


## Attributes Reference

 * `release_tag` - Tag of release
 * `release_id` - ID of release
 * `target_commitish` - Commitish value that determines where the Git release is created from
 * `name` - Name of release
 * `body` - Contents of the description (body) of a release
 * `draft` - (`Boolean`) indicates whether the release is a draft
 * `prerelease` - (`Boolean`) indicates whether the release is a prerelease
 * `created_at` - Date of release creation
 * `published_at` - Date of release publishing
 * `url` - Base URL of the release
 * `html_url` - URL directing to detailed information on the release
 * `asserts_url` - (deprecated) URL of any associated assets with the release
 * `assets_url` - URL of any associated assets with the release
 * `upload_url` - URL that can be used to upload Assets to the release
 * `zipball_url` - Download URL of a specific release in `zip` format
 * `tarball_url` - Download URL of a specific release in `tar.gz` format
