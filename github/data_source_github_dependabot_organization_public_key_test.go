package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubDependabotOrganizationPublicKeyDataSource(t *testing.T) {
	t.Run("queries an organization public key without error", func(t *testing.T) {
		config := `
			data "github_dependabot_organization_public_key" "test" {}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"data.github_dependabot_organization_public_key.test", "key",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
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
