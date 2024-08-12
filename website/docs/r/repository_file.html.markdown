---
layout: "github"
page_title: "GitHub: github_repository_file"
description: |-
  Creates and manages files within a GitHub repository
---

# github_repository_file

This resource allows you to create and manage files within a
GitHub repository.


## Example Usage

### Existing Branch
```hcl

resource "github_repository" "foo" {
  name      = "tf-acc-test-%s"
  auto_init = true
}

resource "github_repository_file" "foo" {
  repository          = github_repository.foo.name
  branch              = "main"
  file                = ".gitignore"
  content             = "**/*.tfstate"
  commit_message      = "Managed by Terraform"
  commit_author       = "Terraform User"
  commit_email        = "terraform@example.com"
  overwrite_on_create = true
}

```

### Auto Created Branch
```hcl

resource "github_repository" "foo" {
  name      = "tf-acc-test-%s"
  auto_init = true
}

resource "github_repository_file" "foo" {
  repository          = github_repository.foo.name
  branch              = "does/not/exist"
  file                = ".gitignore"
  content             = "**/*.tfstate"
  commit_message      = "Managed by Terraform"
  commit_author       = "Terraform User"
  commit_email        = "terraform@example.com"
  overwrite_on_create = true
  autocreate_branch   = true
}

```


## Argument Reference

The following arguments are supported:

* `repository` - (Required) The repository to create the file in.

* `file` - (Required) The path of the file to manage.

* `content` - (Required) The file content.

* `branch` - (Optional) Git branch (defaults to the repository's default branch).
  The branch must already exist, it will only be created automatically if 'autocreate_branch' is set true.

* `commit_author` - (Optional) Committer author name to use. **NOTE:** GitHub app users may omit author and email information so GitHub can verify commits as the GitHub App. This maybe useful when a branch protection rule requires signed commits.

* `commit_email` - (Optional) Committer email address to use. **NOTE:** GitHub app users may omit author and email information so GitHub can verify commits as the GitHub App. This may be useful when a branch protection rule requires signed commits.

* `commit_message` - (Optional) The commit message when creating, updating or deleting the managed file.

* `overwrite_on_create` - (Optional) Enable overwriting existing files. If set to `true` it will overwrite an existing file with the same name. If set to `false` it will fail if there is an existing file with the same name.

* `use_contents_api` - (Optional) Enable to modify github files with the github contents API, rather than git. Defaults to true.

* `pgp_signing_key` - (Optional) PGP signing key used to sign commits. **NOTE:** This must be used with `use_contents_api` set to `false`.

* `pgp_signing_key_passphrase` - (Optional) Passphrase to unlock PGP signing key used to sign commits. **NOTE:** This must be used with `use_contents_api` set to `false`.

* `autocreate_branch` - (Optional) Automatically create the branch if it could not be found. Defaults to false. Subsequent reads if the branch is deleted will occur from 'autocreate_branch_source_branch'.

* `autocreate_branch_source_branch` - (Optional) The branch name to start from, if 'autocreate_branch' is set. Defaults to 'main'.

* `autocreate_branch_source_sha` - (Optional) The commit hash to start from, if 'autocreate_branch' is set. Defaults to the tip of 'autocreate_branch_source_branch'. If provided, 'autocreate_branch_source_branch' is ignored.

## Attributes Reference

The following additional attributes are exported:

* `commit_sha` - The SHA of the commit that modified the file.

* `sha` - The SHA blob of the file.

* `ref` - The name of the commit/branch/tag.


## Import

Repository files can be imported using a combination of the `repo` and `file`, e.g.

```
$ terraform import github_repository_file.gitignore example/.gitignore
```

To import a file from a branch other than the default branch, append `:` and the branch name, e.g.

```
$ terraform import github_repository_file.gitignore example/.gitignore:dev
```
