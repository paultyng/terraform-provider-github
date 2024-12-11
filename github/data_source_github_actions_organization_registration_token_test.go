package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsOrganizationRegistrationTokenDataSource(t *testing.T) {
	t.Run("get an organization registration token without error", func(t *testing.T) {
		config := `
			data "github_actions_organization_registration_token" "test" {
			}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_actions_organization_registration_token.test", "token"),
			resource.TestCheckResourceAttrSet("data.github_actions_organization_registration_token.test", "expires_at"),
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
