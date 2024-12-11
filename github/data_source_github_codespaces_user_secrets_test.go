package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubCodespacesUserSecretsDataSource(t *testing.T) {
	t.Run("queries user codespaces secrets from a repository", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_codespaces_user_secret" "test" {
				secret_name     = "user_cs_secret_1_%s"
				plaintext_value = "foo"
			}
		`, randomID)

		config2 := config + `
			data "github_codespaces_user_secrets" "test" {
			}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_codespaces_user_secrets.test", "secrets.#", "1"),
			resource.TestCheckResourceAttr("data.github_codespaces_user_secrets.test", "secrets.0.name", strings.ToUpper(fmt.Sprintf("USER_CS_SECRET_1_%s", randomID))),
			resource.TestCheckResourceAttrSet("data.github_codespaces_user_secrets.test", "secrets.0.created_at"),
			resource.TestCheckResourceAttrSet("data.github_codespaces_user_secrets.test", "secrets.0.updated_at"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, individual) },
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
