---
layout: "github"
page_title: "GitHub: github_repository_dependabot_security_updates"
description: |-
  Manages automated security fixes for a single repository
---

# github_repository_dependabot_security_updates

This resource allows you to manage dependabot automated security fixes for a single repository. See the 
[documentation](https://docs.github.com/en/code-security/dependabot/dependabot-security-updates/about-dependabot-security-updates)
for details of usage and how this will impact your repository

## Example Usage

```hcl
resource "github_repository" "repo" {
  name         = "my-repo"
  description  = "GitHub repo managed by Terraform"
  
  private = false
  
  vulnerability_alerts   = true
}


resource "github_repository_dependabot_security_updates" "example" {
  repository  = github_repository.test.name
  enabled     = true
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) The name of the GitHub repository.

* `enabled` - (Required) The state of the automated security fixes.

## Import

Automated security references can be imported using the `name` of the repository

### Import by name

```sh
terraform import github_repository_dependabot_security_updates.example my-repo
```
