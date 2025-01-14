package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGithubOrganizationRoles(t *testing.T) {
	t.Run("get the organization roles without error", func(t *testing.T) {
		config := `
			data "github_organization_roles" "test" {}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_roles.test", "roles.#"),
					),
				},
			},
		})
	})
}
