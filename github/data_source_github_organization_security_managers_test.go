package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGithubOrganizationSecurityManagers(t *testing.T) {
	t.Run("get the organization security managers without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		teamName := fmt.Sprintf("tf-acc-%s", randomID)

		config := fmt.Sprintf(`
			resource "github_team" "test" {
				name = "%s"
			}

			resource "github_organization_security_manager" "test" {
				team_slug = github_team.test.slug
			}

			data "github_organization_security_managers" "test" {
				depends_on = [
					github_organization_security_manager.test
				]
			}
		`, teamName)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_security_managers.test", "teams.#"),
						resource.TestCheckResourceAttr("data.github_organization_security_managers.test", "teams.#", "1"),
						resource.TestCheckResourceAttr("data.github_organization_security_managers.test", "teams.0.name", teamName),
					),
				},
			},
		})
	})
}
