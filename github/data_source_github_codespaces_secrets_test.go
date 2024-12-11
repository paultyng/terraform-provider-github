package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubCodespacesSecretsDataSource(t *testing.T) {
	t.Run("queries codespaces secrets from a repository", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_codespaces_secret" "test" {
				secret_name 		= "cs_secret_1"
				repository  		= github_repository.test.name
				plaintext_value = "foo"
			}
		`, randomID)

		config2 := config + `
			data "github_codespaces_secrets" "test" {
				name = github_repository.test.name
			}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_codespaces_secrets.test", "name", fmt.Sprintf("tf-acc-test-%s", randomID)),
			resource.TestCheckResourceAttr("data.github_codespaces_secrets.test", "secrets.#", "1"),
			resource.TestCheckResourceAttr("data.github_codespaces_secrets.test", "secrets.0.name", "CS_SECRET_1"),
			resource.TestCheckResourceAttrSet("data.github_codespaces_secrets.test", "secrets.0.created_at"),
			resource.TestCheckResourceAttrSet("data.github_codespaces_secrets.test", "secrets.0.updated_at"),
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
