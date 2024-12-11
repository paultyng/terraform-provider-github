package github

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubMembership(t *testing.T) {
	if len(testAccConf.testExternalUser) == 0 {
		t.Skip("No external user provided")
	}

	t.Run("creates organization membership", func(t *testing.T) {
		var membership github.Membership
		rn := "github_membership.test_org_membership"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubMembershipDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubMembershipConfig(testAccConf.testExternalUser),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubMembershipExists(rn, &membership),
						testAccCheckGithubMembershipRoleState(rn, &membership),
					),
				},
				{
					ResourceName:      rn,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("creates organization membership with downgrade", func(t *testing.T) {
		var membership github.Membership
		rn := "github_membership.test_org_membership"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubMembershipDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubMembershipConfigDowngradable(testAccConf.testExternalUser),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubMembershipExists(rn, &membership),
						testAccCheckGithubMembershipRoleState(rn, &membership),
					),
				},
				{
					ResourceName: rn,
					ImportState:  true,
				},
			},
		})
	})

	t.Run("creates organization membership with case insensitivity", func(t *testing.T) {
		var membership github.Membership
		var otherMembership github.Membership

		rn := "github_membership.test_org_membership"
		otherCase := flipUsernameCase(testAccConf.testExternalUser)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubMembershipDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubMembershipConfig(testAccConf.testExternalUser),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubMembershipExists(rn, &membership),
					),
				},
				{
					Config: testAccGithubMembershipConfig(otherCase),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubMembershipExists(rn, &otherMembership),
						testAccGithubMembershipTheSame(&membership, &otherMembership),
					),
				},
				{
					ResourceName:      rn,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func testAccCheckGithubMembershipDestroy(s *terraform.State) error {
	meta, err := getTestMeta()
	if err != nil {
		return err
	}
	conn := meta.v3client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_membership" {
			continue
		}

		orgName, username, err := parseTwoPartID(rs.Primary.ID, "organization", "username")
		if err != nil {
			return err
		}

		downgradedOnDestroy := rs.Primary.Attributes["downgrade_on_destroy"] == "true"
		membership, resp, err := conn.Organizations.GetOrgMembership(context.TODO(), username, orgName)
		responseIsSuccessful := err == nil && membership != nil && buildTwoPartID(orgName, username) == rs.Primary.ID

		if downgradedOnDestroy {
			if !responseIsSuccessful {
				return fmt.Errorf("could not load organization membership for %q", rs.Primary.ID)
			}

			if *membership.Role != "member" {
				return fmt.Errorf("organization membership %q is not a member of the org or is not the 'member' role", rs.Primary.ID)
			}

			// Now actually remove them from the org to clean up
			_, removeErr := conn.Organizations.RemoveOrgMembership(context.TODO(), username, orgName)
			if removeErr != nil {
				return fmt.Errorf("organization membership %q could not be removed during membership downgrade test case cleanup: %s", rs.Primary.ID, removeErr)
			}
		} else if responseIsSuccessful {
			return fmt.Errorf("organization membership %q still exists", rs.Primary.ID)
		} else if resp.StatusCode != 404 {
			return err
		}

		return nil
	}
	return nil
}

func testAccCheckGithubMembershipExists(n string, membership *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no membership ID is set")
		}

		meta, err := getTestMeta()
		if err != nil {
			return err
		}
		conn := meta.v3client

		orgName, username, err := parseTwoPartID(rs.Primary.ID, "organization", "username")
		if err != nil {
			return err
		}

		githubMembership, _, err := conn.Organizations.GetOrgMembership(context.TODO(), username, orgName)
		if err != nil {
			return err
		}
		*membership = *githubMembership
		return nil
	}
}

func testAccCheckGithubMembershipRoleState(n string, membership *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no membership ID is set")
		}

		meta, err := getTestMeta()
		if err != nil {
			return err
		}
		conn := meta.v3client

		orgName, username, err := parseTwoPartID(rs.Primary.ID, "organization", "username")
		if err != nil {
			return err
		}

		githubMembership, _, err := conn.Organizations.GetOrgMembership(context.TODO(), username, orgName)
		if err != nil {
			return err
		}

		resourceRole := membership.GetRole()
		actualRole := githubMembership.GetRole()

		if resourceRole != actualRole {
			return fmt.Errorf("membership role %v in resource does match actual state of %v",
				resourceRole, actualRole)
		}
		return nil
	}
}

func testAccGithubMembershipConfig(username string) string {
	return fmt.Sprintf(`
  resource "github_membership" "test_org_membership" {
    username = "%s"
    role = "member"
  }
`, username)
}

func testAccGithubMembershipConfigDowngradable(username string) string {
	return fmt.Sprintf(`
  resource "github_membership" "test_org_membership" {
    username = "%s"
    role = "admin"
    downgrade_on_destroy = %t
  }
`, username, true)
}

func testAccGithubMembershipTheSame(orig, other *github.Membership) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if orig.GetURL() != other.GetURL() {
			return errors.New("users are different")
		}

		return nil
	}
}
