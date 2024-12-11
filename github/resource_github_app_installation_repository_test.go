package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubAppInstallationRepository(t *testing.T) {
	if testAccConf.testOrgAppInstallationId == 0 {
		t.Skip("No org app installation id provided")
	}

	t.Run("installs an app to a repository", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_app_installation_repository" "test" {
				# The installation id of the app (in the organization).
				installation_id    = "%d"
				repository         = github_repository.test.name
			}

		`, randomID, testAccConf.testOrgAppInstallationId)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"github_app_installation_repository.test", "installation_id",
			),
			resource.TestCheckResourceAttrSet(
				"github_app_installation_repository.test", "repo_id",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})
}
