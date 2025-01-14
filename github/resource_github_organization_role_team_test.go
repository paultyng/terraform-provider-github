package github

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationRoleTeam(t *testing.T) {
	t.Run("adds team to an organization org role", func(t *testing.T) {
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
		`, teamName, roleId)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("github_organization_role_team.test", "role_id", strconv.Itoa(roleId)),
						resource.TestCheckResourceAttrPair("github_team.test", "slug", "github_organization_role_team.test", "team_slug"),
					),
				},
			},
		})
	})
}
