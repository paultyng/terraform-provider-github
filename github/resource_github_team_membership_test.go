package github

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubTeamMembership(t *testing.T) {
	if len(testAccConf.testOrgUser) == 0 {
		t.Skip("No test user provided")
	}

	t.Run("creates a team membership", func(t *testing.T) {
		var membership github.Membership
		randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubTeamMembershipDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubTeamMembershipConfig(randString, testAccConf.testOrgUser, "member"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubTeamMembershipExists("github_team_membership.test_team_membership", &membership),
						testAccCheckGithubTeamMembershipRoleState("github_team_membership.test_team_membership", "member", &membership),
						testAccCheckGithubTeamMembershipExists("github_team_membership.test_team_membership_slug", &membership),
						testAccCheckGithubTeamMembershipRoleState("github_team_membership.test_team_membership_slug", "member", &membership),
					),
				},
				{
					Config: testAccGithubTeamMembershipConfig(randString, testAccConf.testOrgUser, "maintainer"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubTeamMembershipExists("github_team_membership.test_team_membership", &membership),
						testAccCheckGithubTeamMembershipRoleState("github_team_membership.test_team_membership", "maintainer", &membership),
						testAccCheckGithubTeamMembershipExists("github_team_membership.test_team_membership_slug", &membership),
						testAccCheckGithubTeamMembershipRoleState("github_team_membership.test_team_membership_slug", "maintainer", &membership),
					),
				},
			},
		})
	})

	t.Run("is case insensitive", func(t *testing.T) {
		var membership github.Membership
		var otherMembership github.Membership

		rn := "github_team_membership.test_team_membership"
		randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

		otherCase := flipUsernameCase(testAccConf.testOrgUser)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubTeamMembershipDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubTeamMembershipConfig(randString, testAccConf.testOrgUser, "member"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubTeamMembershipExists(rn, &membership),
					),
				},
				{
					Config: testAccGithubTeamMembershipConfig(randString, otherCase, "member"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubTeamMembershipExists(rn, &otherMembership),
						testAccGithubTeamMembershipTheSame(&membership, &otherMembership),
					),
				},
			},
		})
	})
}

func testAccCheckGithubTeamMembershipDestroy(s *terraform.State) error {
	meta, err := getTestMeta()
	if err != nil {
		return err
	}
	conn := meta.v3client
	orgId := meta.id

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_team_membership" {
			continue
		}

		teamIdString, username, err := parseTwoPartID(rs.Primary.ID, "team_id", "username")
		if err != nil {
			return err
		}

		teamId, err := getTeamID(teamIdString, meta)
		if err != nil {
			return unconvertibleIdErr(teamIdString, err)
		}

		membership, resp, err := conn.Teams.GetTeamMembershipByID(context.TODO(),
			orgId, teamId, username)
		if err == nil {
			if membership != nil {
				return fmt.Errorf("team membership still exists")
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckGithubTeamMembershipExists(n string, membership *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no team membership ID is set")
		}

		meta, err := getTestMeta()
		if err != nil {
			return err
		}
		conn := meta.v3client
		orgId := meta.id
		teamIdString, username, err := parseTwoPartID(rs.Primary.ID, "team_id", "username")
		if err != nil {
			return err
		}

		teamId, err := getTeamID(teamIdString, meta)
		if err != nil {
			return unconvertibleIdErr(teamIdString, err)
		}

		teamMembership, _, err := conn.Teams.GetTeamMembershipByID(context.TODO(), orgId, teamId, username)
		if err != nil {
			return err
		}
		*membership = *teamMembership
		return nil
	}
}

func testAccCheckGithubTeamMembershipRoleState(n, expected string, membership *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no team membership ID is set")
		}

		meta, err := getTestMeta()
		if err != nil {
			return err
		}
		conn := meta.v3client
		orgId := meta.id
		teamIdString, username, err := parseTwoPartID(rs.Primary.ID, "team_id", "username")
		if err != nil {
			return err
		}
		teamId, err := getTeamID(teamIdString, meta)
		if err != nil {
			return unconvertibleIdErr(teamIdString, err)
		}

		teamMembership, _, err := conn.Teams.GetTeamMembershipByID(context.TODO(),
			orgId, teamId, username)
		if err != nil {
			return err
		}

		resourceRole := membership.GetRole()
		actualRole := teamMembership.GetRole()

		if resourceRole != expected {
			return fmt.Errorf("team membership role %v in resource does match expected state of %v", resourceRole, expected)
		}

		if resourceRole != actualRole {
			return fmt.Errorf("team membership role %v in resource does match actual state of %v", resourceRole, actualRole)
		}
		return nil
	}
}

func testAccGithubTeamMembershipConfig(randString, username, role string) string {
	return fmt.Sprintf(`
resource "github_team" "test_team" {
  name        = "tf-acc-test-team-membership-%s"
  description = "Terraform acc test group"
}

resource "github_team" "test_team_slug" {
  name        = "tf-acc-test-team-membership-%s-slug"
  description = "Terraform acc test group"
}

resource "github_team_membership" "test_team_membership" {
  team_id  = github_team.test_team.id
  username = "%s"
  role     = "%s"
}

resource "github_team_membership" "test_team_membership_slug" {
  team_id  = github_team.test_team_slug.slug
  username = "%s"
  role     = "%s"
}
`, randString, randString, username, role, username, role)
}

func testAccGithubTeamMembershipTheSame(orig, other *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *orig.URL != *other.URL {
			return errors.New("users are different")
		}

		return nil
	}
}
