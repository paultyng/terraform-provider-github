package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubBranch(t *testing.T) {
	t.Run("creates a branch directly", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
		resource "github_repository" "test" {
			name = "tf-acc-test-%[1]s"
			auto_init = true
		}

		resource "github_branch" "test" {
			repository = github_repository.test.name
			branch     = "test"
		}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_branch.test", "id"),
						resource.TestCheckResourceAttr("github_branch.test", "ref", "refs/heads/test"),
						resource.TestCheckResourceAttrSet("github_branch.test", "sha"),
					),
				},
			},
		})
	})

	t.Run("creates a branch named main directly and a repository with a gitignore_template", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
		resource "github_repository" "test" {
			name = "tf-acc-test-%[1]s"
			auto_init = true
			gitignore_template = "Python"
		}

		resource "github_branch" "test" {
			repository = github_repository.test.id
			branch     = "main"
		}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_branch.test", "id"),
						resource.TestCheckResourceAttr("github_branch.test", "ref", "refs/heads/main"),
						resource.TestCheckResourceAttrSet("github_branch.test", "sha"),
					),
				},
			},
		})
	})

	t.Run("creates a branch from a source branch", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
		resource "github_repository" "test" {
			name = "tf-acc-test-%[1]s"
			auto_init = true
		}

		resource "github_branch" "source" {
			repository = github_repository.test.id
			branch     = "source"
		}

		resource "github_branch" "test" {
			repository    = github_repository.test.id
			source_branch = github_branch.source.branch
			branch        = "test"
		}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_branch.test", "id"),
						resource.TestCheckResourceAttr("github_branch.test", "ref", "refs/heads/test"),
						resource.TestCheckResourceAttrPair("github_branch.test", "sha", "github_branch.source", "sha"),
					),
				},
			},
		})
	})
}
