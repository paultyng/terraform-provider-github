package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationRepositoryRole(t *testing.T) {
	t.Run("can create an organization repository role without erroring", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("tf-acc-org-repo-role-%s", randomID)
		description := "This is a test org repo role."
		baseRole := "write"
		permission0 := "reopen_issue"
		permission1 := "reopen_pull_request"

		config := fmt.Sprintf(`
			resource "github_organization_repository_role" "test" {
				name        = "%s"
				description = "%s"
				base_role   = "%s"
				permissions = [
				"%s",
				"%s"
				]
			}
		`, name, description, baseRole, permission0, permission1)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, enterprise) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_organization_repository_role.test", "id"),
						resource.TestCheckResourceAttrSet("github_organization_repository_role.test", "role_id"),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "name", name),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "description", description),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "base_role", baseRole),
						resource.TestCheckResourceAttrSet("github_organization_repository_role.test", "permissions.#"),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "permissions.#", "2"),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "permissions.0", permission0),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "permissions.1", permission1),
					),
				},
			},
		})
	})

	t.Run("can create an minimal organization repository role without erroring", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("tf-acc-org-repo-role-%s", randomID)
		permission0 := "reopen_issue"

		config := fmt.Sprintf(`
			resource "github_organization_repository_role" "test" {
				name        = "%s"
				permissions = [
				"%s"
				]
			}
		`, name, permission0)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, enterprise) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_organization_repository_role.test", "id"),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "name", name),
						resource.TestCheckResourceAttrSet("github_organization_repository_role.test", "base_role"),
						resource.TestCheckResourceAttrSet("github_organization_repository_role.test", "permissions.#"),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "permissions.#", "1"),
						resource.TestCheckResourceAttr("github_organization_repository_role.test", "permissions.0", permission0),
					),
				},
			},
		})
	})
}
