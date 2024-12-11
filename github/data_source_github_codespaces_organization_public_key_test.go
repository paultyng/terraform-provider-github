package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubCodespacesOrganizationPublicKeyDataSource(t *testing.T) {
	t.Run("queries an organization public key without error", func(t *testing.T) {
		config := `
			data "github_codespaces_organization_public_key" "test" {}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_codespaces_organization_public_key.test", "key_id"),
						resource.TestCheckResourceAttrSet("data.github_codespaces_organization_public_key.test", "key"),
					),
				},
			},
		})
	})
}
