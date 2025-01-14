package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationRepositoryRoleDataSource(t *testing.T) {
	t.Run("queries an organization repository role", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		roleName := fmt.Sprintf(`tf-acc-test-%s`, randomID)

		config := fmt.Sprintf(`
			resource "github_organization_repository_role" "test" {
				name        = "%s"
				description = "Test role description"
				base_role   = "read"
				permissions = [
					"reopen_issue",
					"reopen_pull_request",
				]
			}

			data "github_organization_repository_role" "test" {
				role_id = github_organization_repository_role.test.role_id

				depends_on = [ github_organization_repository_role.test ]
			}
		`, roleName)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, enterprise) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("data.github_organization_repository_role.test", "name", "github_organization_repository_role.test", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_repository_role.test", "description", "github_organization_repository_role.test", "description"),
						resource.TestCheckResourceAttrPair("data.github_organization_repository_role.test", "base_role", "github_organization_repository_role.test", "base_role"),
						resource.TestCheckResourceAttrPair("data.github_organization_repository_role.test", "permissions.#", "github_organization_repository_role.test", "permissions.#"),
					),
				},
			},
		})
	})
}
