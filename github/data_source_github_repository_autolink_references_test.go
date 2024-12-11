package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryAutolinkReferencesDataSource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("queries autolink references", func(t *testing.T) {
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%[1]s"
				auto_init = true
			}

			resource "github_repository_autolink_reference" "autolink_default" {
				repository = github_repository.test.name

				key_prefix          = "TEST1-"
				target_url_template = "https://example.com/TEST-<num>"
			}
		`, randomID)

		config2 := config + `
			data "github_repository_autolink_references" "all" {
				repository = github_repository.test.name
			}
		`
		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_repository_autolink_references.all", "autolink_references.#", "1"),
			resource.TestCheckResourceAttr("data.github_repository_autolink_references.all", "autolink_references.0.key_prefix", "TEST1-"),
			resource.TestCheckResourceAttr("data.github_repository_autolink_references.all", "autolink_references.0.target_url_template", "https://example.com/TEST-<num>"),
			resource.TestCheckResourceAttr("data.github_repository_autolink_references.all", "autolink_references.0.is_alphanumeric", "true"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					Config: config2,
					Check:  check,
				},
			},
		})
	})
}
