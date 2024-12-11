package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationTeamsDataSource(t *testing.T) {
	t.Run("queries no org teams without error", func(t *testing.T) {
		config := `
			data "github_organization_teams" "all" {}
		`

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.#"),
			resource.TestCheckResourceAttr("data.github_organization_teams.all", "teams.#", "0"),
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

	t.Run("queries without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		teamName0 := fmt.Sprintf("tf-acc-test-0-%s", randomID)
		teamName1 := fmt.Sprintf("tf-acc-test-1-%s", randomID)
		teamName2 := fmt.Sprintf("tf-acc-test-2-%s", randomID)

		config0 := fmt.Sprintf(`
			resource "github_team" "test_0" {
				name = "%s"
			}

			resource "github_team" "test_1" {
				name    = "%s"
				privacy = "closed"
			}

			resource "github_team" "test_2" {
				name           = "%s"
				privacy        = "closed"
				parent_team_id = github_team.test_1.id
			}

			data "github_organization_teams" "all" {
				depends_on = [github_team.test_0, github_team.test_1, github_team.test_2]
			}
		`, teamName0, teamName1, teamName2)

		config1 := fmt.Sprintf(`
			resource "github_team" "test_0" {
				name = "%s"
			}

			resource "github_team" "test_1" {
				name    = "%s"
				privacy = "closed"
			}

			resource "github_team" "test_2" {
				name           = "%s"
				privacy        = "closed"
				parent_team_id = github_team.test_1.id
			}

			data "github_organization_teams" "all" {
			  root_teams_only = true

				depends_on = [github_team.test_0, github_team.test_1, github_team.test_2]
			}
		`, teamName0, teamName1, teamName2)

		config2 := fmt.Sprintf(`
			resource "github_team" "test_0" {
				name = "%s"
			}

			resource "github_team" "test_1" {
				name    = "%s"
				privacy = "closed"
			}

			resource "github_team" "test_2" {
				name           = "%s"
				privacy        = "closed"
				parent_team_id = github_team.test_1.id
			}

			data "github_organization_teams" "all" {
			  summary_only = true

				depends_on = [github_team.test_0, github_team.test_1, github_team.test_2]
			}
		`, teamName0, teamName1, teamName2)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config0,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.#"),
						resource.TestCheckResourceAttr("data.github_organization_teams.all", "teams.#", "3"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.node_id"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.slug", "github_team.test_0", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.name", "github_team.test_0", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.description", "github_team.test_0", "description"),
						// resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.privacy", "github_team.test_0", "privacy"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.parent_team_id", "github_team.test_0", "parent_team_id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.members.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.repositories.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.node_id"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.slug", "github_team.test_1", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.name", "github_team.test_1", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.description", "github_team.test_1", "description"),
						// resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.privacy", "github_team.test_1", "privacy"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.parent_team_id", "github_team.test_1", "parent_team_id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.members.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.repositories.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.2.id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.2.node_id"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.slug", "github_team.test_2", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.name", "github_team.test_2", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.description", "github_team.test_2", "description"),
						// resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.privacy", "github_team.test_2", "privacy"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.parent_team_id", "github_team.test_2", "parent_team_id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.2.members.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.2.repositories.#"),
					),
				},
				{
					Config: config1,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.#"),
						resource.TestCheckResourceAttr("data.github_organization_teams.all", "teams.#", "2"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.node_id"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.slug", "github_team.test_0", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.name", "github_team.test_0", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.description", "github_team.test_0", "description"),
						// resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.privacy", "github_team.test_0", "privacy"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.parent_team_id", "github_team.test_0", "parent_team_id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.members.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.repositories.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.node_id"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.slug", "github_team.test_1", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.name", "github_team.test_1", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.description", "github_team.test_1", "description"),
						// resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.privacy", "github_team.test_1", "privacy"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.parent_team_id", "github_team.test_1", "parent_team_id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.members.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.repositories.#"),
					),
				},
				{
					Config: config2,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.#"),
						resource.TestCheckResourceAttr("data.github_organization_teams.all", "teams.#", "3"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.0.node_id"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.slug", "github_team.test_0", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.name", "github_team.test_0", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.description", "github_team.test_0", "description"),
						// resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.privacy", "github_team.test_0", "privacy"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.0.parent_team_id", "github_team.test_0", "parent_team_id"),
						resource.TestCheckNoResourceAttr("data.github_organization_teams.all", "teams.0.members.#"),
						resource.TestCheckNoResourceAttr("data.github_organization_teams.all", "teams.0.repositories.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.1.node_id"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.slug", "github_team.test_1", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.name", "github_team.test_1", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.description", "github_team.test_1", "description"),
						// resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.privacy", "github_team.test_1", "privacy"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.1.parent_team_id", "github_team.test_1", "parent_team_id"),
						resource.TestCheckNoResourceAttr("data.github_organization_teams.all", "teams.1.members.#"),
						resource.TestCheckNoResourceAttr("data.github_organization_teams.all", "teams.1.repositories.#"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.2.id"),
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.2.node_id"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.slug", "github_team.test_2", "slug"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.name", "github_team.test_2", "name"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.description", "github_team.test_2", "description"),
						// resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.privacy", "github_team.test_2", "privacy"),
						resource.TestCheckResourceAttrPair("data.github_organization_teams.all", "teams.2.parent_team_id", "github_team.test_2", "parent_team_id"),
						resource.TestCheckNoResourceAttr("data.github_organization_teams.all", "teams.2.members.#"),
						resource.TestCheckNoResourceAttr("data.github_organization_teams.all", "teams.2.repositories.#"),
					),
				},
			},
		})
	})

	t.Run("queries results_per_page only without error", func(t *testing.T) {
		config := `
		data "github_organization_teams" "all" {
			results_per_page = 50
		}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_teams.all", "teams.#"),
						resource.TestCheckResourceAttr("data.github_organization_teams.all", "teams.#", "0"),
					),
				},
			},
		})
	})
}
