package github

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubTeamMembers(t *testing.T) {
	if len(testAccConf.testOrgUser) == 0 {
		t.Skip("No test user provided")
	}

	t.Run("creates a team & members configured with defaults", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		teamName := fmt.Sprintf("tf-acc-test-team-%s", randomID)

		config := fmt.Sprintf(`
resource "github_team" "test" {
  name        = "%s"
}

resource "github_team_members" "test" {
  team_id  = github_team.test.id

	members {
		username = "%s"
		role     = "member"
	}
}
`, teamName, testAccConf.testOrgUser)

		configUpdated := fmt.Sprintf(`
resource "github_team" "test" {
  name        = "%s"
}

resource "github_team_members" "test" {
  team_id  = github_team.test.id

	members {
		username = "%s"
		role     = "maintainer"
	}
}
`, teamName, testAccConf.testOrgUser)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubTeamMembersDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_team_members.test", "members.#"),
						resource.TestCheckResourceAttr("github_team_members.test", "members.#", "1"),
						resource.TestCheckResourceAttr("github_team_members.test", "members.0.username", testAccConf.testOrgUser),
						resource.TestCheckResourceAttr("github_team_members.test", "members.0.role", "member"),
					),
				},
				{
					Config: configUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_team_members.test", "members.#"),
						resource.TestCheckResourceAttr("github_team_members.test", "members.#", "1"),
						resource.TestCheckResourceAttr("github_team_members.test", "members.0.username", testAccConf.testOrgUser),
						resource.TestCheckResourceAttr("github_team_members.test", "members.0.role", "member"),
					),
				},
			},
		})
	})
}

func testAccCheckGithubTeamMembersDestroy(s *terraform.State) error {
	meta, err := getTestMeta()
	if err != nil {
		return err
	}
	conn := meta.v3client
	orgId := meta.id

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_team_members" {
			continue
		}

		teamIdString := rs.Primary.ID

		teamId, err := getTeamID(teamIdString, meta)
		if err != nil {
			return unconvertibleIdErr(teamIdString, err)
		}

		members, resp, err := conn.Teams.ListTeamMembersByID(context.TODO(),
			orgId, teamId, nil)
		if err == nil {
			if len(members) > 0 {
				return fmt.Errorf("team has still members: %v", members)
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}
