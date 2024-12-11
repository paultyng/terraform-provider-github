package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsOrganizationOIDCSubjectClaimCustomizationTemplateDataSource(t *testing.T) {
	t.Run("get an organization oidc subject claim customization template without error", func(t *testing.T) {
		config := `
			resource "github_actions_organization_oidc_subject_claim_customization_template" "test" {
				include_claim_keys = ["actor", "actor_id", "head_ref", "repository"]
			}
		`

		config2 := config + `
			data "github_actions_organization_oidc_subject_claim_customization_template" "test" {}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_actions_organization_oidc_subject_claim_customization_template.test", "include_claim_keys.#", "4"),
			resource.TestCheckResourceAttr("data.github_actions_organization_oidc_subject_claim_customization_template.test", "include_claim_keys.0", "actor"),
			resource.TestCheckResourceAttr("data.github_actions_organization_oidc_subject_claim_customization_template.test", "include_claim_keys.1", "actor_id"),
			resource.TestCheckResourceAttr("data.github_actions_organization_oidc_subject_claim_customization_template.test", "include_claim_keys.2", "head_ref"),
			resource.TestCheckResourceAttr("data.github_actions_organization_oidc_subject_claim_customization_template.test", "include_claim_keys.3", "repository"),
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
