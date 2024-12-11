package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationSecurityManagers(t *testing.T) {
	t.Run("adds team as security manager", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_team" "test" {
				name = "tf-acc-%s"
			}

			resource "github_organization_security_manager" "test" {
				team_slug = github_team.test.slug
			}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("github_team.test", "ID", "github_organization_security_manager.test", "ID"),
						resource.TestCheckResourceAttrPair("github_team.test", "slug", "github_organization_security_manager.test", "team_slug"),
						resource.TestCheckResourceAttr("github_organization_security_manager.test", "team_slug", fmt.Sprintf("tf-acc-%s", randomID)),
					),
				},
			},
		})
	})

	t.Run("handles team name changes", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_team" "test" {
				name = "tf-acc-%s"
			}

			resource "github_organization_security_manager" "test" {
				team_slug = github_team.test.slug
			}
		`, randomID)

		configUpdated := fmt.Sprintf(`
			resource "github_team" "test" {
				name = "tf-acc-updated-%s"
			}

			resource "github_organization_security_manager" "test" {
				team_slug = github_team.test.slug
			}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("github_team.test", "ID", "github_organization_security_manager.test", "ID"),
						resource.TestCheckResourceAttrPair("github_team.test", "slug", "github_organization_security_manager.test", "team_slug"),
						resource.TestCheckResourceAttr("github_organization_security_manager.test", "team_slug", fmt.Sprintf("tf-acc-%s", randomID)),
					),
				},
				{
					Config: configUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("github_team.test", "ID", "github_organization_security_manager.test", "ID"),
						resource.TestCheckResourceAttrPair("github_team.test", "slug", "github_organization_security_manager.test", "team_slug"),
						resource.TestCheckResourceAttr("github_organization_security_manager.test", "team_slug", fmt.Sprintf("tf-acc-updated-%s", randomID)),
					),
				},
			},
		})
	})

	t.Run("handles team name changes", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_team" "test" {
				name = "tf-acc-%s"
			}

			resource "github_organization_security_manager" "test" {
				team_slug = github_team.test.slug
			}
		`, randomID)

		configUpdated := fmt.Sprintf(`
			resource "github_team" "test" {
				name = "tf-acc-updated-%s"
			}

			resource "github_organization_security_manager" "test" {
				team_slug = github_team.test.slug
			}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("github_team.test", "ID", "github_organization_security_manager.test", "ID"),
						resource.TestCheckResourceAttrPair("github_team.test", "slug", "github_organization_security_manager.test", "team_slug"),
						resource.TestCheckResourceAttr("github_organization_security_manager.test", "team_slug", fmt.Sprintf("tf-acc-%s", randomID)),
					),
				},
				{
					Config: configUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("github_team.test", "ID", "github_organization_security_manager.test", "ID"),
						resource.TestCheckResourceAttrPair("github_team.test", "slug", "github_organization_security_manager.test", "team_slug"),
						resource.TestCheckResourceAttr("github_organization_security_manager.test", "team_slug", fmt.Sprintf("tf-acc-updated-%s", randomID)),
					),
				},
			},
		})
	})
}
