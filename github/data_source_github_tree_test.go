package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubTreeDataSource(t *testing.T) {
	t.Run("get tree", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "this" {
				auto_init = true
				name      = "tf-acc-test-%s"
			}

			data "github_branch" "this" {
				branch     = "main"
				repository = github_repository.this.name
			}

			data "github_tree" "this" {
				recursive  = false
				repository = github_repository.this.name
				tree_sha   = data.github_branch.this.sha
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"data.github_tree.this", "entries.#",
			),
			resource.TestCheckResourceAttr(
				"data.github_tree.this", "entries.0.path",
				"README.md",
			),
			resource.TestCheckResourceAttr(
				"data.github_tree.this", "entries.0.type",
				"blob",
			),
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
