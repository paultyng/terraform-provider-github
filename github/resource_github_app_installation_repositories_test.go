package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubAppInstallationRepositories(t *testing.T) {
	if testAccConf.testOrgAppInstallationId == 0 {
		t.Skip("No org app installation id provided")
	}

	t.Run("installs an app to multiple repositories", func(t *testing.T) {
		randomId := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName0 := fmt.Sprintf("tf-acc-test-0-%s", randomId)
		repoName1 := fmt.Sprintf("tf-acc-test-1-%s", randomId)

		config := fmt.Sprintf(`
		resource "github_repository" "test_0" {
			name      = "%s"
			auto_init = true
		}

		resource "github_repository" "test_1" {
			name      = "tf-acc-test-%s"
			auto_init = true
		}

		resource "github_app_installation_repositories" "test" {
			# The installation id of the app (in the organization).
			installation_id         = "%d"
			selected_repositories = [github_repository.test_0.name, github_repository.test_1.name]
		}

		`, repoName0, repoName1, testAccConf.testOrgAppInstallationId)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"github_app_installation_repositories.test", "installation_id",
			),
			resource.TestCheckResourceAttr(
				"github_app_installation_repositories.test", "selected_repositories.#", "2",
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
