package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryDeploymentBranchPolicy(t *testing.T) {
	t.Run("creates deployment branch policy", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_repository_environment" "env" {
				repository  = github_repository.test.name
				environment = "my_env"
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}
		`, randomID)

		config1 := `
			resource "github_repository_deployment_branch_policy" "br" {
				repository       = github_repository.test.name
				environment_name = github_repository_environment.env.environment
				name             = github_repository.test.default_branch
			}
		`

		config2 := `
			resource "github_repository_deployment_branch_policy" "br" {
				repository       = github_repository.test.name
				environment_name = github_repository_environment.env.environment
				name             = "foo"
			}
		`

		check1 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_deployment_branch_policy.br", "name", "main",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_deployment_branch_policy.br", "etag",
			),
		)

		check2 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_deployment_branch_policy.br", "name", "foo",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config + config1,
					Check:  check1,
				},
				{
					Config: config + config2,
					Check:  check2,
				},
			},
		})
	})
}
