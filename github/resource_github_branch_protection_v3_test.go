package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubBranchProtectionV3(t *testing.T) {
	t.Run("configures default settings when empty", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
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

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "branch", "main"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "require_signed_commits", "false"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "require_conversation_resolution", "false"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_status_checks.#", "0"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_pull_request_reviews.#", "0"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "push_restrictions.#", "0"),
					),
				},
			},
		})
	})

	t.Run("configures conversation resolution", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
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

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "branch", "main"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "require_signed_commits", "false"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "require_conversation_resolution", "true"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_status_checks.#", "0"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_pull_request_reviews.#", "0"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "push_restrictions.#", "0"),
					),
				},
			},
		})
	})

	t.Run("configures required status checks", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
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
				checks = [
					"github/foo",
					"github/bar:-1",
					"github:foo:baz:1",
				]
			}

		}
	`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_branch_protection_v3.test", "required_status_checks.#"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_status_checks.#", "1"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_status_checks.0.strict", "true"),
						resource.TestCheckResourceAttrSet("github_branch_protection_v3.test", "required_status_checks.0.checks.#"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_status_checks.0.checks.#", "3"),
					),
				},
			},
		})
	})

	t.Run("configures required status checks context", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
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

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_branch_protection_v3.test", "required_status_checks.#"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_status_checks.#", "1"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_status_checks.0.strict", "true"),
						resource.TestCheckResourceAttrSet("github_branch_protection_v3.test", "required_status_checks.0.checks.#"),
						resource.TestCheckResourceAttr("github_branch_protection_v3.test", "required_status_checks.0.checks.#", "1"),
					),
				},
			},
		})
	})

	t.Run("configures required pull request reviews", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_team" "test" {
				name = "tf-acc-test-%[1]s"
			}

			resource "github_team_repository" "test" {
				team_id    = github_team.test.id
				repository = github_repository.test.name
				permission = "admin"
			}

			resource "github_branch_protection_v3" "test" {

			  repository  = github_repository.test.name
			  branch      = "main"

			  required_pull_request_reviews {
				  dismiss_stale_reviews      = true
				  require_code_owner_reviews = true
				  required_approving_review_count = 1
				  require_last_push_approval = true
				  dismissal_users = ["a"]
				  dismissal_teams = ["b"]
				  dismissal_apps = ["c"]
				  bypass_pull_request_allowances {
					  users = ["d"]
					  teams = [github_team.test.slug]
					  apps = ["e"]
				  }
			  }

			  depends_on = [github_team_repository.test]
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
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.require_last_push_approval", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.dismissal_users.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.dismissal_teams.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.dismissal_apps.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.bypass_pull_request_allowances.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.bypass_pull_request_allowances.0.users.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.bypass_pull_request_allowances.0.teams.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.bypass_pull_request_allowances.0.apps.#", "1",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("configures required pull request reviews with bypass allowances", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_team" "test" {
				name = "tf-acc-test-%[1]s"
			}

			resource "github_team_repository" "test" {
				team_id    = github_team.test.id
				repository = github_repository.test.name
				permission = "admin"
			}

			resource "github_branch_protection_v3" "test" {
			  repository  = github_repository.test.name
			  branch      = "main"

			  required_pull_request_reviews {
					bypass_pull_request_allowances {
						teams = [github_team.test.slug]
					}
			  }

				depends_on = [github_team_repository.test]
			}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection_v3.test", "required_pull_request_reviews.0.bypass_pull_request_allowances.#", "1",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("configures branch push restrictions", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
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

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
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
