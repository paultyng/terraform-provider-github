package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryPullRequest(t *testing.T) {
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
		`, randomID)

		const resourceName = "github_repository_pull_request.test"

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				resourceName, "base_repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(resourceName, "base_ref", "main"),
			resource.TestCheckResourceAttr(resourceName, "head_ref", "test"),
			resource.TestCheckResourceAttr(resourceName, "title", "test title"),
			resource.TestCheckResourceAttr(resourceName, "body", "test body"),
			resource.TestCheckResourceAttr(resourceName, "maintainer_can_modify", "false"),
			resource.TestCheckResourceAttrSet(resourceName, "base_sha"),
			resource.TestCheckResourceAttr(resourceName, "draft", "false"),
			resource.TestCheckResourceAttrSet(resourceName, "head_sha"),
			resource.TestCheckResourceAttr(resourceName, "labels.#", "0"),
			resource.TestCheckResourceAttrSet(resourceName, "number"),
			resource.TestCheckResourceAttrSet(resourceName, "opened_at"),
			resource.TestCheckResourceAttrSet(resourceName, "opened_by"),
			resource.TestCheckResourceAttr(resourceName, "state", "open"),
			resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
