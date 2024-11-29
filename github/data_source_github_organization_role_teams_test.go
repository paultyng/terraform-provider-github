package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGithubOrganizationRoleTeams(t *testing.T) {
	t.Run("get the organization role teams without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		teamName := fmt.Sprintf("tf-acc-team-%s", randomID)
		roleId := 8134
		config := fmt.Sprintf(`
			resource "github_team" "test" {
				name = "%s"
			}

			resource "github_organization_role_team" "test" {
				role_id   = %d
				team_slug = github_team.test.slug
			}

			data "github_organization_role_teams" "test" {
				role_id = %[2]d

				depends_on = [
					github_organization_role_team.test
				]
			}
		`, teamName, roleId)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_role_teams.test", "teams.#"),
						resource.TestCheckResourceAttr("data.github_organization_role_teams.test", "teams.#", "1"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.0.team_id", "github_team.test", "id"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.0.slug", "github_team.test", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.0.name", "github_team.test", "name"),
					),
				},
			},
		})
	})

	t.Run("get indirect organization role teams without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		teamName1 := fmt.Sprintf("tf-acc-team-1-%s", randomID)
		teamName2 := fmt.Sprintf("tf-acc-team-2-%s", randomID)
		roleId := 8134
		config := fmt.Sprintf(`
			resource "github_team" "test_1" {
				name    = "%s"
				privacy = "closed"
			}

			resource "github_team" "test_2" {
				name           = "%s"
				privacy        = "closed"
				parent_team_id = github_team.test_1.id
			}

			resource "github_organization_role_team" "test" {
				role_id   = %d
				team_slug = github_team.test_1.slug
			}

			data "github_organization_role_teams" "test" {
				role_id = %[3]d

				depends_on = [
					github_organization_role_team.test
				]
			}
		`, teamName1, teamName2, roleId)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_role_teams.test", "teams.#"),
						resource.TestCheckResourceAttr("data.github_organization_role_teams.test", "teams.#", "2"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.0.team_id", "github_team.test_1", "id"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.0.slug", "github_team.test_1", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.0.name", "github_team.test_1", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.1.team_id", "github_team.test_2", "id"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.1.slug", "github_team.test_2", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_role_teams.test", "teams.1.name", "github_team.test_2", "name"),
					),
				},
			},
		})
	})
}
