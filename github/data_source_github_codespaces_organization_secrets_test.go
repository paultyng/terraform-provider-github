package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubCodespacesOrganizationSecretsDataSource(t *testing.T) {
	t.Run("queries organization codespaces secrets from a repository", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		secretName := fmt.Sprintf("ORG_CS_SECRET_1_%s", randomID)

		config := fmt.Sprintf(`
		resource "github_codespaces_organization_secret" "test" {
			secret_name 		= %s"
			plaintext_value = "foo"
			visibility      = "private"
		}

		data "github_codespaces_organization_secrets" "test" {
		  depends_on = [github_codespaces_organization_secret.test]
		}
		`, secretName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.github_codespaces_organization_secrets.test", "secrets.#", "1"),
						resource.TestCheckResourceAttr("data.github_codespaces_organization_secrets.test", "secrets.0.name", secretName),
						resource.TestCheckResourceAttr("data.github_codespaces_organization_secrets.test", "secrets.0.visibility", "private"),
						resource.TestCheckResourceAttrSet("data.github_codespaces_organization_secrets.test", "secrets.0.created_at"),
						resource.TestCheckResourceAttrSet("data.github_codespaces_organization_secrets.test", "secrets.0.updated_at"),
					),
				},
			},
		})
	})
}
