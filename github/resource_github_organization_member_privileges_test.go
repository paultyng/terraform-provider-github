package github

import (
	"fmt"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccGithubOrganizationMemberPrivileges_basic(t *testing.T) {
	if err := testAccCheckOrganization(); err != nil {
		t.Skipf("Skipping because %s.", err.Error())
	}

	var testOrganization github.Organization

	testCase := func(t *testing.T, mode string) {
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, mode) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubOrganizationMemberPrivilegesConfigBefore,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubOrganizationMemberPrivilegesAttributes(&testOrganization, &testAccGithubOrganizationMemberPrivilegesExpectedAttributes{
							DefaultRepoPermission: "none",
							MembersCanCreateRepos: false,
						}),
					),
				},
				{
					Config: testAccGithubOrganizationMemberPrivilegesConfigAfter,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubOrganizationMemberPrivilegesAttributes(&testOrganization, &testAccGithubOrganizationMemberPrivilegesExpectedAttributes{
							DefaultRepoPermission: "read",
							MembersCanCreateRepos: true,
						}),
					),
				},
				{
					ResourceName:      "github_organization_member_privileges.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	}

	t.Run("with an anonymous account", func(t *testing.T) {
		t.Skip("anonymous account not supported for this operation")
	})

	t.Run("with an individual account", func(t *testing.T) {
		t.Skip("individual account not supported for this operation")
	})

	t.Run("with an organization account", func(t *testing.T) {
		testCase(t, organization)
	})
}

type testAccGithubOrganizationMemberPrivilegesExpectedAttributes struct {
	DefaultRepoPermission string
	MembersCanCreateRepos bool
}

func testAccCheckGithubOrganizationMemberPrivilegesAttributes(organization *github.Organization, want *testAccGithubOrganizationMemberPrivilegesExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if defaultRepoPermission := organization.GetDefaultRepoPermission(); defaultRepoPermission != want.DefaultRepoPermission {
			return fmt.Errorf("got defaultRepoPermission %q; want %q", defaultRepoPermission, want.DefaultRepoPermission)
		}
		if membersCanCreateRepos := organization.GetMembersCanCreateRepos(); membersCanCreateRepos != want.MembersCanCreateRepos {
			return fmt.Errorf("got membersCanCreateRepos %t; want %t", membersCanCreateRepos, want.MembersCanCreateRepos)
		}

		return nil
	}
}

const testAccGithubOrganizationMemberPrivilegesConfigBefore = `
resource "github_organization_member_privileges" "test" {
	default_repository_permission   = "none"
	members_can_create_repositories = false
}
`

const testAccGithubOrganizationMemberPrivilegesConfigAfter = `
resource "github_organization_member_privileges" "test" {
	default_repository_permission   = "read"
	members_can_create_repositories = true
}
`
