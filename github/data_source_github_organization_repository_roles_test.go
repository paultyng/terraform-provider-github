package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGithubOrganizationRepositoryRoles(t *testing.T) {
	t.Run("get empty organization roles without error", func(t *testing.T) {
		config := `
			data "github_organization_repository_roles" "test" {}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_repository_roles.test", "roles.#"),
						resource.TestCheckResourceAttr("data.github_organization_repository_roles.test", "roles.#", "0"),
					),
				},
			},
		})
	})

	t.Run("get organization roles without error", func(t *testing.T) {
		config := `
		resource "github_organization_repository_role" "test" {
				name        = "%s"
				description = "Test role description"
				base_role   = "read"
				permissions = [
					"reopen_issue",
					"reopen_pull_request",
				]
			}

			data "github_organization_repository_roles" "test" {
				depends_on = [ github_organization_repository_role.test ]
			}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, enterprise) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_repository_roles.test", "roles.#"),
						resource.TestCheckResourceAttr("data.github_organization_repository_roles.test", "roles.#", "1"),
						resource.TestCheckResourceAttrPair("data.github_organization_repository_roles.test", "roles.0.name", "github_organization_repository_role.test", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_repository_roles.test", "roles.0.description", "github_organization_repository_role.test", "description"),
						resource.TestCheckResourceAttrPair("data.github_organization_repository_roles.test", "roles.0.base_role", "github_organization_repository_role.test", "base_role"),
						resource.TestCheckResourceAttrPair("data.github_organization_repository_roles.test", "roles.0.permissions.#", "github_organization_repository_role.test", "permissions.#"),
					),
				},
			},
		})
	})
}
