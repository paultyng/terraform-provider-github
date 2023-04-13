---
layout: "github"
page_title: "GitHub: github_repository"
description: |-
  Get details about GitHub repository
---

# github_repository

Use this data source to retrieve information about a GitHub repository.

## Example Usage

```hcl
data "github_repository" "example" {
  full_name = "hashicorp/terraform"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the repository.

* `full_name` - (Optional) Full name of the repository (in `org/name` format).

## Attributes Reference

* `node_id` - the Node ID of the repository.

* `description` - A description of the repository.

* `homepage_url` - URL of a page describing the project.

* `private` - Whether the repository is private.

* `visibility` - Whether the repository is public, private or internal.

* `has_issues` - Whether the repository has GitHub Issues enabled.

* `has_discussions` - Whether the repository has GitHub Discussions enabled.

* `has_projects` - Whether the repository has the GitHub Projects enabled.

* `has_wiki` - Whether the repository has the GitHub Wiki enabled.

* `is_template` - Whether the repository is a template repository.

* `fork` - Whether the repository is a fork.

* `allow_merge_commit` - Whether the repository allows merge commits.

* `allow_squash_merge` - Whether the repository allows squash merges.

* `allow_rebase_merge` - Whether the repository allows rebase merges.

* `allow_auto_merge` - Whether the repository allows auto-merging pull requests.

* `allow_forking` - Whether the repository allows private forks.

* `squash_merge_commit_title` - The default value for a squash merge commit title.

* `squash_merge_commit_message` - The default value for a squash merge commit message.

* `merge_commit_title` - The default value for a merge commit title.

* `merge_commit_message` - The default value for a merge commit message.

* `has_downloads` - Whether the repository has Downloads feature enabled.

* `default_branch` - The name of the default branch of the repository.

* `primary_language` - The primary language used in the repository.

* `archived` - Whether the repository is archived.

* `pages` - The repository's GitHub Pages configuration.

* `topics` - The list of topics of the repository.

* `template` - The repository source template configuration.

* `html_url` - URL to the repository on the web.

* `ssh_clone_url` - URL that can be provided to `git clone` to clone the repository via SSH.

* `http_clone_url` - URL that can be provided to `git clone` to clone the repository via HTTPS.

* `git_clone_url` - URL that can be provided to `git clone` to clone the repository anonymously via the git protocol.

* `svn_url` - URL that can be provided to `svn checkout` to check out the repository via GitHub's Subversion protocol emulation.

* `node_id` - GraphQL global node id for use with v4 API

* `repo_id` - GitHub ID for the repository

* `repository_license` - An Array of GitHub repository licenses. Each `repository_license` block consists of the fields documented below.

___

The `repository_license` block consists of:

* `content` - Content of the license file, encoded by encoding scheme mentioned below.
* `download_url` - The URL to download the raw content of the license file.
* `encoding` - The encoding used for the content (e.g., "base64").
* `git_url` - The URL to access information about the license file as a Git blob.
* `html_url` - The URL to view the license file on GitHub.
* `license` - `license` block consists of the fields documented below.
* `name` - The name of the license file (e.g., "LICENSE").
* `path` - The path to the license file within the repository.
* `sha` - The SHA hash of the license file.
* `size` - The size of the license file in bytes.
* `type` - The type of the content, (e.g., "file").
* `url` - The URL to access information about the license file on GitHub.

The `license` block consists of:

* `body` - The text of the license.
* `conditions` - Conditions associated with the license.
* `description` - A description of the license.
* `featured` - Indicates if the license is featured.
* `html_url` - The URL to view the license details on GitHub.
* `implementation` - Details about the implementation of the license.
* `key` - A key representing the license type (e.g., "apache-2.0").
* `limitations` - Limitations associated with the license.
* `name` - The name of the license (e.g., "Apache License 2.0").
* `permissions` - Permissions associated with the license.
* `spdx_id` - The SPDX identifier for the license (e.g., "Apache-2.0").
* `url` - The URL to access information about the license on GitHub.
