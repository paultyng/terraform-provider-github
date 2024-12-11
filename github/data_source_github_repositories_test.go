package github

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoriesDataSource(t *testing.T) {
	// FIXME: Find a way to reduce amount of `GET /search/repositories`
	// t.Skip("Skipping due to API rate limits exceeding")

	t.Run("queries a list of repositories without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_repositories" "test" {
				query = "org:%s"
			}
		`, testAccConf.owner)

		check := resource.ComposeTestCheckFunc(
			resource.TestMatchResourceAttr(
				"data.github_repositories.test", "full_names.0",
				regexp.MustCompile(`^`+testAccConf.owner),
			),
			resource.TestCheckResourceAttrSet(
				"data.github_repositories.test", "names.0",
			),
			resource.TestCheckNoResourceAttr(
				"data.github_repositories.test", "repo_ids.0",
			),
			resource.TestCheckResourceAttr(
				"data.github_repositories.test", "sort",
				"updated",
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

	t.Run("queries a list of repositories with repo_ids and results_per_page without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_repositories" "test" {
				query = "org:%s"
				include_repo_id = true
				results_per_page = 20
			}
		`, testAccConf.owner)

		check := resource.ComposeTestCheckFunc(
			resource.TestMatchResourceAttr(
				"data.github_repositories.test", "full_names.0",
				regexp.MustCompile(`^`+testAccConf.owner),
			),
			resource.TestCheckResourceAttrSet(
				"data.github_repositories.test", "names.0",
			),
			resource.TestCheckResourceAttrSet(
				"data.github_repositories.test", "repo_ids.0",
			),
			resource.TestCheckResourceAttr(
				"data.github_repositories.test", "sort",
				"updated",
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

	t.Run("returns an empty list given an invalid query", func(t *testing.T) {
		// FIXME: Find a way to reduce amount of `GET /search/repositories`
		// t.Skip("Skipping due to API rate limits exceeding")

		config := `
			data "github_repositories" "test" {
				query = "klsafj_23434_doesnt_exist"
			}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repositories.test", "full_names.#",
				"0",
			),
			resource.TestCheckResourceAttr(
				"data.github_repositories.test", "names.#",
				"0",
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
