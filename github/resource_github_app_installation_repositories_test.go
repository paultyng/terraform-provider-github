package github

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubAppInstallationRepositories(t *testing.T) {

	const APP_INSTALLATION_ID = "APP_INSTALLATION_ID"
	randomID1 := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	randomID2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	installation_id, exists := os.LookupEnv(APP_INSTALLATION_ID)

	t.Run("installs an app to multiple repositories", func(t *testing.T) {

		if !exists {
			t.Skipf("%s environment variable is missing", APP_INSTALLATION_ID)
		}

		config := fmt.Sprintf(`

		resource "github_repository" "test1" {
			name      = "tf-acc-test-%s"
			auto_init = true
		}

		resource "github_repository" "test2" {
			name      = "tf-acc-test-%s"
			auto_init = true
		}

		resource "github_app_installation_repositories" "test" {
			# The installation id of the app (in the organization).
			installation_id         = "%s"
			selected_repositories = [github_repository.test1.name, github_repository.test2.name]
		}

		`, randomID1, randomID2, installation_id)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"github_app_installation_repositories.test", "installation_id",
			),
			resource.TestCheckResourceAttr(
				"github_app_installation_repositories.test", "selected_repositories.#", "2",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

}
