package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationExternalIdentities(t *testing.T) {
	t.Run("queries without error", func(t *testing.T) {
		config := `data "github_organization_external_identities" "test" {}`

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_organization_external_identities.test", "identities.#"),
			resource.TestCheckResourceAttrSet("data.github_organization_external_identities.test", "identities.0.login"),
			resource.TestCheckResourceAttrSet("data.github_organization_external_identities.test", "identities.0.saml_identity.name_id"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
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
