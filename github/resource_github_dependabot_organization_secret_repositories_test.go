package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubDependabotOrganizationSecretRepositories(t *testing.T) {
	t.Run("set repository allowlist for an organization secret", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
		resource "github_actions_organization_secret" "test" {
			secret_name     = "TEST"
			plaintext_value = "Testing 1..2..3.."
			visibility      = "all"
		}

		resource "github_repository" "test_repo_1" {
			name = "tf-acc-test-%s-1"
			visibility = "private"
			vulnerability_alerts = "true"
		}

		resource "github_repository" "test_repo_2" {
			name = "tf-acc-test-%s-2"
			visibility = "private"
			vulnerability_alerts = "true"
		}

		resource "github_dependabot_organization_secret_repositories" "org_secret_repos" {
			secret_name = github_actions_organization_secret.test.secret_name
			selected_repository_ids = [
				github_repository.test_repo_1.repo_id,
				github_repository.test_repo_2.repo_id
			]
		}
		`, randomID, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_dependabot_organization_secret_repositories.org_secret_repos", "secret_name"),
						resource.TestCheckResourceAttr("github_dependabot_organization_secret_repositories.org_secret_repos", "selected_repository_ids.#", "2"),
					),
				},
			},
		})
	})
}
