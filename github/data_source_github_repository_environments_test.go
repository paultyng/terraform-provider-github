package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryEnvironmentsDataSource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("queries environments", func(t *testing.T) {
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%[1]s"
				auto_init = true
			}
			resource "github_repository_environment" "env1" {
				repository = github_repository.test.name
				environment = "env_x"
			}
		`, randomID)

		config2 := config + `
			data "github_repository_environments" "all" {
				repository = github_repository.test.name
			}
		`
		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_repository_environments.all", "environments.#", "1"),
			resource.TestCheckResourceAttr("data.github_repository_environments.all", "environments.0.name", "env_x"),
			resource.TestCheckResourceAttrSet("data.github_repository_environments.all", "environments.0.node_id"),
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
