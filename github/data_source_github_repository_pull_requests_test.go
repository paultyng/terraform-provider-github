package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryPullRequestsDataSource(t *testing.T) {
	t.Run("manages the pull request lifecycle", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_branch" "test" {
				repository    = github_repository.test.name
				branch        = "test"
				source_branch = github_repository.test.default_branch
			}

			resource "github_repository_file" "test" {
				repository     = github_repository.test.name
				branch         = github_branch.test.branch
				file           = "test"
				content        = "bar"
			}

			resource "github_repository_pull_request" "test" {
				base_repository = github_repository_file.test.repository
				base_ref        = github_repository.test.default_branch
				head_ref        = github_branch.test.branch
				title           = "test title"
				body            = "test body"
			}

			data "github_repository_pull_requests" "test" {
				base_repository = github_repository_pull_request.test.base_repository
				head_ref        = github_branch.test.branch
				base_ref        = github_repository.test.default_branch
				sort_by         = "updated"
				sort_direction  = "desc"
				state           = "open"
			}
		`, randomID)

		const resourceName = "data.github_repository_pull_requests.test"

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "base_repository", fmt.Sprintf("tf-acc-test-%s", randomID)),
			resource.TestCheckResourceAttr(resourceName, "state", "open"),
			resource.TestCheckResourceAttr(resourceName, "base_ref", "main"),
			resource.TestCheckResourceAttr(resourceName, "head_ref", "test"),
			resource.TestCheckResourceAttr(resourceName, "sort_by", "updated"),
			resource.TestCheckResourceAttr(resourceName, "sort_direction", "desc"),

			resource.TestCheckResourceAttr(resourceName, "results.#", "1"),
			resource.TestCheckResourceAttrSet(resourceName, "results.0.number"),
			resource.TestCheckResourceAttr(resourceName, "results.0.base_ref", "main"),
			resource.TestCheckResourceAttrSet(resourceName, "results.0.base_sha"),
			resource.TestCheckResourceAttr(resourceName, "results.0.body", "test body"),
			resource.TestCheckResourceAttr(resourceName, "results.0.draft", "false"),
			resource.TestCheckResourceAttrSet(resourceName, "results.0.head_owner"),
			resource.TestCheckResourceAttr(resourceName, "results.0.head_ref", "test"),
			resource.TestCheckResourceAttr(resourceName, "results.0.head_repository", fmt.Sprintf("tf-acc-test-%s", randomID)),
			resource.TestCheckResourceAttrSet(resourceName, "results.0.head_sha"),
			resource.TestCheckResourceAttr(resourceName, "results.0.labels.#", "0"),
			resource.TestCheckResourceAttr(resourceName, "results.0.maintainer_can_modify", "false"),
			resource.TestCheckResourceAttrSet(resourceName, "results.0.opened_at"),
			resource.TestCheckResourceAttrSet(resourceName, "results.0.opened_by"),
			resource.TestCheckResourceAttr(resourceName, "results.0.state", "open"),
			resource.TestCheckResourceAttr(resourceName, "results.0.title", "test title"),
			resource.TestCheckResourceAttrSet(resourceName, "results.0.updated_at"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})
}
