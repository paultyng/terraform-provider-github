package github

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccGithubTeamMembership_basic(t *testing.T) {
	if testCollaborator == "" {
		t.Skip("Skipping because `GITHUB_TEST_COLLABORATOR` is not set")
	}
	if err := testAccCheckOrganization(); err != nil {
		t.Skipf("Skipping because %s.", err.Error())
	}

	var membership github.Membership

	rn := "github_team_membership.test_team_membership"
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGithubTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubTeamMembershipConfig(randString, testCollaborator, "member"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubTeamMembershipExists(rn, &membership),
					testAccCheckGithubTeamMembershipRoleState(rn, "member", &membership),
				),
			},
			{
				Config: testAccGithubTeamMembershipConfig(randString, testCollaborator, "maintainer"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubTeamMembershipExists(rn, &membership),
					testAccCheckGithubTeamMembershipRoleState(rn, "maintainer", &membership),
				),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGithubTeamMembership_caseInsensitive(t *testing.T) {
	if testCollaborator == "" {
		t.Skip("Skipping because `GITHUB_TEST_COLLABORATOR` is not set")
	}
	if err := testAccCheckOrganization(); err != nil {
		t.Skipf("Skipping because %s.", err.Error())
	}

	var membership github.Membership
	var otherMembership github.Membership

	rn := "github_team_membership.test_team_membership"
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	otherCase := flipUsernameCase(testCollaborator)

	if testCollaborator == otherCase {
		t.Skip("Skipping because `GITHUB_TEST_COLLABORATOR` has no letters to flip case")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGithubTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubTeamMembershipConfig(randString, testCollaborator, "member"),
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
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGithubTeamMembershipDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Owner).v3client
	orgId := testAccProvider.Meta().(*Owner).id

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_team_membership" {
			continue
		}

		teamIdString, username, err := parseTwoPartID(rs.Primary.ID, "team_id", "username")
		if err != nil {
			return err
		}

		teamId, err := strconv.ParseInt(teamIdString, 10, 64)
		if err != nil {
			return unconvertibleIdErr(teamIdString, err)
		}

		membership, resp, err := conn.Teams.GetTeamMembershipByID(context.TODO(),
			orgId, teamId, username)
		if err == nil {
			if membership != nil {
				return fmt.Errorf("Team membership still exists")
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
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No team membership ID is set")
		}

		conn := testAccProvider.Meta().(*Owner).v3client
		orgId := testAccProvider.Meta().(*Owner).id
		teamIdString, username, err := parseTwoPartID(rs.Primary.ID, "team_id", "username")
		if err != nil {
			return err
		}

		teamId, err := strconv.ParseInt(teamIdString, 10, 64)
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
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No team membership ID is set")
		}

		conn := testAccProvider.Meta().(*Owner).v3client
		orgId := testAccProvider.Meta().(*Owner).id
		teamIdString, username, err := parseTwoPartID(rs.Primary.ID, "team_id", "username")
		if err != nil {
			return err
		}
		teamId, err := strconv.ParseInt(teamIdString, 10, 64)
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
			return fmt.Errorf("Team membership role %v in resource does match expected state of %v", resourceRole, expected)
		}

		if resourceRole != actualRole {
			return fmt.Errorf("Team membership role %v in resource does match actual state of %v", resourceRole, actualRole)
		}
		return nil
	}
}

func testAccGithubTeamMembershipConfig(randString, username, role string) string {
	return fmt.Sprintf(`
resource "github_membership" "test_org_membership" {
  username = "%s"
  role     = "member"
}

resource "github_team" "test_team" {
  name        = "tf-acc-test-team-membership-%s"
  description = "Terraform acc test group"
}

resource "github_team_membership" "test_team_membership" {
  team_id  = "${github_team.test_team.id}"
  username = "%s"
  role     = "%s"
}
`, username, randString, username, role)
}

func testAccGithubTeamMembershipTheSame(orig, other *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *orig.URL != *other.URL {
			return errors.New("users are different")
		}

		return nil
	}
}
