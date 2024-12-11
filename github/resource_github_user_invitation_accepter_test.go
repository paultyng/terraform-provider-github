package github

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubUserInvitationAccepter(t *testing.T) {
	if len(testAccConf.testExternalUser) == 0 {
		t.Skip("No external user provided")
	}

	if len(testAccConf.testExternalUserToken) == 0 {
		t.Skip("No external user token provided")
	}

	t.Run("accepts an invitation", func(t *testing.T) {
		rn := "github_repository_collaborator.test"
		repoName := fmt.Sprintf("tf-acc-test-collab-%s", acctest.RandString(5))

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubUserInvitationAccepterDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubUserInvitationAccepterConfig(testAccConf.testExternalUserToken, repoName, testAccConf.testExternalUser),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(rn, "permission", "push"),
						resource.TestMatchResourceAttr(rn, "invitation_id", regexp.MustCompile(`^[0-9]+$`)),
					),
				},
			},
		})
	})

	t.Run("accepts an invitation with an empty invitation_id", func(t *testing.T) {
		rn := "github_user_invitation_accepter.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubUserInvitationAccepterDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubUserInvitationAccepterAllowEmptyId(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(rn, "invitation_id", ""),
						resource.TestCheckResourceAttr(rn, "allow_empty_id", "true"),
					),
				},
			},
		})
	})
}

func testAccCheckGithubUserInvitationAccepterDestroy(s *terraform.State) error {
	return nil
}

func testAccGithubUserInvitationAccepterConfig(inviteeToken, repoName, collaborator string) string {
	return fmt.Sprintf(`
provider "github" {
  alias = "main"
}

provider "github" {
  alias = "invitee"
  token = "%s"
}

resource "github_repository" "test" {
  provider = "github.main"
  name     = "%s"
}

resource "github_repository_collaborator" "test" {
  provider   = "github.main"
  repository = "${github_repository.test.name}"
  username   = "%s"
  permission = "push"
}

resource "github_user_invitation_accepter" "test" {
  provider      = "github.invitee"
  invitation_id = "${github_repository_collaborator.test.invitation_id}"
}
`, inviteeToken, repoName, collaborator)
}

func testAccGithubUserInvitationAccepterAllowEmptyId() string {
	return `
provider "github" {}

resource "github_user_invitation_accepter" "test" {
  invitation_id  = ""
  allow_empty_id = true
}
`
}
