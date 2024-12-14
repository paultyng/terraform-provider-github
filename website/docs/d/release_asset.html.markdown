---
layout: "github"
page_title: "GitHub: github_release_asset"
description: |-
  Get information on a GitHub release asset.
---

# github\_release\_asset

Use this data source to retrieve information about a GitHub release asset associated with a specific GitHub release.

## Example Usage
To retrieve the latest release that is present in a repository:

```hcl
data "github_release" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    retrieve_by = "latest"
}
```

To retrieve a specific release asset from a repository based on its ID:

```hcl
data "github_release_asset" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    asset_id    = 12345
}
```

To retrieve the first release asset associated with the the latest release in a repository:

```hcl
data "github_release" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    retrieve_by = "latest"
}

data "github_release_asset" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    asset_id    = data.github_release.example.assets[0].id
}
```

To retrieve all release assets associated with the the latest release in a repository:

```hcl
data "github_release" "example" {
    repository  = "example-repository"
    owner       = "example-owner"
    retrieve_by = "latest"
}

data "github_release_asset" "example" {
    count       = length(data.github_release.example.assets)
    repository  = "example-repository"
    owner       = "example-owner"
    asset_id    = data.github_release.example.assets[count.index].id
}
```

## Argument Reference

*  `repository`  -  (Required) Name of the repository to retrieve the release from.
*  `owner`  -  (Required) Owner of the repository.
*  `asset_id`  -  (Required) ID of the release asset to retrieve.

## Attributes Reference

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
* `body` - The release asset body
