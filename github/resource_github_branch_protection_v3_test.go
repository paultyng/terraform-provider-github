package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccGithubBranchProtectionV3_defaults(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("configures default settings when empty", func(t *testing.T) {

		config := fmt.Sprintf(`

		resource "github_repository" "test" {
		  name      = "tf-acc-test-%s"
		  auto_init = true
		}

		resource "github_branch_protection_v3" "test" {

		  repository  = github_repository.test.name
		  branch      = "main"

		}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "branch", "main",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "require_signed_commits", "false",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "require_conversation_resolution", "false",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_status_checks.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "push_restrictions.#", "0",
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
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})
}

func TestAccGithubBranchProtectionV3_conversation_resolution(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("configures default settings when empty", func(t *testing.T) {

		config := fmt.Sprintf(`

		resource "github_repository" "test" {
		  name      = "tf-acc-test-%s"
		  auto_init = true
		}

		resource "github_branch_protection_v3" "test" {

		  repository  = github_repository.test.name
		  branch      = "main"

		  require_conversation_resolution = true
		}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "branch", "main",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "require_signed_commits", "false",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "require_conversation_resolution", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_status_checks.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "push_restrictions.#", "0",
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
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})
}

func TestAccGithubBranchProtectionV3_required_status_checks(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("configures required status checks", func(t *testing.T) {

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection_v3" "test" {

			  repository  = github_repository.test.name
			  branch      = "main"

			  required_status_checks {
			    strict   = true
			    contexts = ["github/foo"]
			  }

			}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_status_checks.#", "1",
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
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})
}
func TestAccGithubBranchProtectionV3_required_pull_request_reviews(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("configures required pull request reviews", func(t *testing.T) {

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection_v3" "test" {

			  repository  = github_repository.test.name
			  branch      = "main"

			  required_pull_request_reviews {
				dismiss_stale_reviews      = true
				require_code_owner_reviews = true
			  }

			}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.dismiss_stale_reviews", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.require_code_owner_reviews", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.required_approving_review_count", "1",
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
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})
}

func TestAccGithubBranchProtectionV3_branch_push_restrictions(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("configures branch push restrictions", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_team" "test" {
				name = "tf-acc-test-%[1]s"
			}
			  
			resource "github_team_repository" "test" {
				team_id    = "${github_team.test.id}"
				repository = "${github_repository.test.name}"
				permission = "pull"
			}

			resource "github_branch_protection_v3" "test" {

			  repository   = github_repository.test.name
			  branch       = "main"

			  restrictions {
				teams = ["${github_team.test.slug}"]
			  }
			  
			}
			`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "restrictions.#", "1",
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
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

}
