package github

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGithubOrganizationRoleUsers(t *testing.T) {
	login := os.Getenv("GITHUB_IN_ORG_USER")
	if len(login) == 0 {
		t.Skip("set inOrgUser to unskip this test run")
	}

	t.Run("get the organization role users without error", func(t *testing.T) {
		roleId := 8134
		config := fmt.Sprintf(`
			resource "github_organization_role_user" "test" {
				role_id = %d
				login   = "%s"
			}

			data "github_organization_role_users" "test" {
				role_id = %[1]d

				depends_on = [
					github_organization_role_user.test
				]
			}
		`, roleId, login)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_role_users.test", "users.#"),
						resource.TestCheckResourceAttr("data.github_organization_role_users.test", "users.#", "1"),
						resource.TestCheckResourceAttrSet("data.github_organization_role_users.test", "users.0.user_id"),
						resource.TestCheckResourceAttr("data.github_organization_role_users.test", "users.0.login", login),
					),
				},
			},
		})
	})

	t.Run("get indirect organization role users without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		teamName := fmt.Sprintf("tf-acc-team-%s", randomID)
		roleId := 8134
		config := fmt.Sprintf(`
			resource "github_team" "test" {
				name    = "%s"
				privacy = "closed"
			}

			resource "github_team_membership" "test" {
				team_id  = github_team.test.id
				username = "%s"
			}

			resource "github_organization_role_team" "test" {
				role_id   = %d
				team_slug = github_team.test.slug
			}

			data "github_organization_role_users" "test" {
				role_id = %[3]d

				depends_on = [
					github_organization_role_team.test
				]
			}
		`, teamName, login, roleId)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_role_users.test", "users.#"),
						resource.TestCheckResourceAttr("data.github_organization_role_users.test", "users.#", "1"),
						resource.TestCheckResourceAttrSet("data.github_organization_role_users.test", "users.0.user_id"),
						resource.TestCheckResourceAttr("data.github_organization_role_users.test", "users.0.login", login),
					),
				},
			},
		})
	})
}
