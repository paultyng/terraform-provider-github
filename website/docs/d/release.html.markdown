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

To retrieve a specific release from a repository based on its ID:

```hcl
data "github_release" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    retrieve_by = "id"
    id          = 12345
}
```

Finally, to retrieve a release based on its tag:

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
 * `assets_url` - URL of any associated assets with the release
 * `asserts_url` - **Deprecated**: Use `assets_url` resource instead
 * `upload_url` - URL that can be used to upload Assets to the release
 * `zipball_url` - Download URL of a specific release in `zip` format
 * `tarball_url` - Download URL of a specific release in `tar.gz` format
 * `assets` - Collection of assets for the release. Each asset conforms to the following schema:
    * `id` - ID of the asset
    * `url` - URL of the asset
    * `node_id` - Node ID of the asset
    * `name` - The file name of the asset
    * `label` - Label for the asset
    * `content_type` - MIME type of the asset
    * `size` - Size in byte
    * `created_at` - Date the asset was created
    * `updated_at` - Date the asset was last updated
    * `browser_download_url` - Browser download URL

