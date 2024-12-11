package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubDependabotOrganizationSecretsDataSource(t *testing.T) {
	t.Run("queries organization dependabot secrets from a repository", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_dependabot_organization_secret" "test" {
				secret_name 		= "org_dep_secret_1_%s"
				plaintext_value = "foo"
				visibility      = "private"
			}
		`, randomID)

		config2 := config + `
			data "github_dependabot_organization_secrets" "test" {
			}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_dependabot_organization_secrets.test", "secrets.#", "1"),
			resource.TestCheckResourceAttr("data.github_dependabot_organization_secrets.test", "secrets.0.name", strings.ToUpper(fmt.Sprintf("ORG_DEP_SECRET_1_%s", randomID))),
			resource.TestCheckResourceAttr("data.github_dependabot_organization_secrets.test", "secrets.0.visibility", "private"),
			resource.TestCheckResourceAttrSet("data.github_dependabot_organization_secrets.test", "secrets.0.created_at"),
			resource.TestCheckResourceAttrSet("data.github_dependabot_organization_secrets.test", "secrets.0.updated_at"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
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
