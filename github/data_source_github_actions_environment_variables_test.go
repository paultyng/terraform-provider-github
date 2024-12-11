package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsEnvironmentVariablesDataSource(t *testing.T) {
	t.Run("queries actions variables from an environment", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			}

			resource "github_repository_environment" "test" {
			  repository       = github_repository.test.name
			  environment      = "environment / test"
			}

			resource "github_actions_environment_variable" "variable" {
			  repository       = github_repository.test.name
			  environment      = github_repository_environment.test.environment
			  variable_name    = "test_variable"
			  value  		   = "foo"
			}
			`, randomID)

		config2 := config + `
			data "github_actions_environment_variables" "test" {
				environment      = github_repository_environment.test.environment
				name 		     = github_repository.test.name
			}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_actions_environment_variables.test", "variables.#", "1"),
			resource.TestCheckResourceAttr("data.github_actions_environment_variables.test", "variables.0.name", strings.ToUpper("test_variable")),
			resource.TestCheckResourceAttr("data.github_actions_environment_variables.test", "variables.0.value", "foo"),
			resource.TestCheckResourceAttrSet("data.github_actions_environment_variables.test", "variables.0.created_at"),
			resource.TestCheckResourceAttrSet("data.github_actions_environment_variables.test", "variables.0.updated_at"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  resource.ComposeTestCheckFunc(),
				},
				{
					Config: config2,
					Check:  check,
				},
			},
		})
	})
}
