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

func TestAccGithubTeamMemberships_basic(t *testing.T) {
	if testCollaborator == "" {
		t.Skip("Skipping because `GITHUB_TEST_COLLABORATOR` is not set")
	}
	if err := testAccCheckOrganization(); err != nil {
		t.Skipf("Skipping because %s.", err.Error())
	}

	var membership github.Membership

	rn := "github_team_membership.test_team_memberships"
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGithubTeamMembershipsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubTeamMembershipsConfig(randString, testCollaborator, "member"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubTeamMembershipsExists(rn, &membership),
					testAccCheckGithubTeamMembershipsRoleState(rn, "member", &membership),
				),
			},
			{
				Config: testAccGithubTeamMembershipsConfig(randString, testCollaborator, "maintainer"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubTeamMembershipsExists(rn, &membership),
					testAccCheckGithubTeamMembershipsRoleState(rn, "maintainer", &membership),
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

func TestAccGithubTeamMemberships_caseInsensitive(t *testing.T) {
	if testCollaborator == "" {
		t.Skip("Skipping because `GITHUB_TEST_COLLABORATOR` is not set")
	}
	if err := testAccCheckOrganization(); err != nil {
		t.Skipf("Skipping because %s.", err.Error())
	}

	var membership github.Membership
	var otherMembership github.Membership

	rn := "github_team_membership.test_team_memberships"
	randString := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	otherCase := flipUsernameCase(testCollaborator)

	if testCollaborator == otherCase {
		t.Skip("Skipping because `GITHUB_TEST_COLLABORATOR` has no letters to flip case")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGithubTeamMembershipsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubTeamMembershipsConfig(randString, testCollaborator, "member"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubTeamMembershipsExists(rn, &membership),
				),
			},
			{
				Config: testAccGithubTeamMembershipsConfig(randString, otherCase, "member"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubTeamMembershipsExists(rn, &otherMembership),
					testAccGithubTeamMembershipsTheSame(&membership, &otherMembership),
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

func testAccCheckGithubTeamMembershipsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Owner).v3client
	orgId := testAccProvider.Meta().(*Owner).id

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_team_memberships" {
			continue
		}

		teamIdString := rs.Primary.ID

		teamId, err := strconv.ParseInt(teamIdString, 10, 64)
		if err != nil {
			return unconvertibleIdErr(teamIdString, err)
		}

		members, resp, err := conn.Teams.ListTeamMembersByID(context.TODO(),
			orgId, teamId, nil)
		if err == nil {
			if len(members) > 0 {
				return fmt.Errorf("Team has still members: %v", members)
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckGithubTeamMembershipsExists(n string, membership *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No team ID is set")
		}

		conn := testAccProvider.Meta().(*Owner).v3client
		orgId := testAccProvider.Meta().(*Owner).id
		teamIdString := rs.Primary.ID

		teamId, err := strconv.ParseInt(teamIdString, 10, 64)
		if err != nil {
			return unconvertibleIdErr(teamIdString, err)
		}

		members, _, err := conn.Teams.ListTeamMembersByID(context.TODO(), orgId, teamId, nil)
		if err != nil {
			return err
		}

		if len(members) != 1 {
			return fmt.Errorf("Team has not one member: %d", len(members))
		}

		TeamMembership, _, err := conn.Teams.GetTeamMembershipByID(context.TODO(), orgId, teamId, *members[0].Login)

		if err != nil {
			return err
		}
		*membership = *TeamMembership
		return nil
	}
}

func testAccCheckGithubTeamMembershipsRoleState(n, expected string, membership *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No team ID is set")
		}

		conn := testAccProvider.Meta().(*Owner).v3client
		orgId := testAccProvider.Meta().(*Owner).id
		teamIdString := rs.Primary.ID

		teamId, err := strconv.ParseInt(teamIdString, 10, 64)
		if err != nil {
			return unconvertibleIdErr(teamIdString, err)
		}

		members, _, err := conn.Teams.ListTeamMembersByID(context.TODO(), orgId, teamId, nil)
		if err != nil {
			return err
		}

		if len(members) != 1 {
			return fmt.Errorf("Team has not one member: %d", len(members))
		}

		TeamMemberships, _, err := conn.Teams.GetTeamMembershipByID(context.TODO(),
			orgId, teamId, *members[0].Login)
		if err != nil {
			return err
		}

		resourceRole := membership.GetRole()
		actualRole := TeamMemberships.GetRole()

		if resourceRole != expected {
			return fmt.Errorf("Team membership role %v in resource does match expected state of %v", resourceRole, expected)
		}

		if resourceRole != actualRole {
			return fmt.Errorf("Team membership role %v in resource does match actual state of %v", resourceRole, actualRole)
		}
		return nil
	}
}

func testAccGithubTeamMembershipsConfig(randString, username, role string) string {
	return fmt.Sprintf(`
resource "github_membership" "test_org_membership" {
  username = "%s"
  role     = "member"
}

resource "github_team" "test_team" {
  name        = "tf-acc-test-team-membership-%s"
  description = "Terraform acc test group"
}

resource "github_team_memberships" "test_team_memberships" {
  team_id  = "${github_team.test_team.id}"
	members {
		username = "%s"
		role     = "%s"
	}
}
`, username, randString, username, role)
}

func testAccGithubTeamMembershipsTheSame(orig, other *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *orig.URL != *other.URL {
			return errors.New("users are different")
		}

		return nil
	}
}
