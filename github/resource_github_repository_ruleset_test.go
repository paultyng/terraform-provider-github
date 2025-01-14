package github

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestGithubRepositoryRulesets(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("Creates and updates repository rulesets without errors", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = false
				vulnerability_alerts = true
			}

			resource "github_repository_environment" "example" {
				environment  = "test"
				repository   = github_repository.test.name
			}

			resource "github_repository_ruleset" "test" {
				name        = "test"
				repository  = github_repository.test.id
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~ALL"]
						exclude = []
					}
				}

				rules {
					creation = true

					update = true

					deletion                = true
					required_linear_history = true

					required_deployments {
						required_deployment_environments = ["test"]
					}

					required_signatures = false

					pull_request {
						required_approving_review_count   = 2
						required_review_thread_resolution = true
						require_code_owner_review         = true
						dismiss_stale_reviews_on_push     = true
						require_last_push_approval        = true
					}

					required_status_checks {

						required_check {
							context = "ci"
						}

						strict_required_status_checks_policy = true
						do_not_enforce_on_create             = true
					}

					required_code_scanning {
					  required_code_scanning_tool {
						alerts_threshold = "errors"
						security_alerts_threshold = "high_or_higher"
						tool = "CodeQL"
					  }
					}

					non_fast_forward = true
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test",
				"name",
				"test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test",
				"enforcement",
				"active",
			),
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test",
				"rules.0.required_code_scanning.0.required_code_scanning_tool.0.alerts_threshold",
				"errors",
			),
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test",
				"rules.0.required_code_scanning.0.required_code_scanning_tool.0.security_alerts_threshold",
				"high_or_higher",
			),
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test",
				"rules.0.required_code_scanning.0.required_code_scanning_tool.0.tool",
				"CodeQL",
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

	t.Run("Creates and updates repository rulesets with enterprise features without errors", func(t *testing.T) {
		if isEnterprise != "true" {
			t.Skip("Skipping because `ENTERPRISE_ACCOUNT` is not set or set to false")
		}

		if testEnterprise == "" {
			t.Skip("Skipping because `ENTERPRISE_SLUG` is not set")
		}

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = false
				vulnerability_alerts = true
			}

			resource "github_repository_environment" "example" {
				environment  = "test"
				repository   = github_repository.test.name
			}

			resource "github_repository_ruleset" "test" {
				name        = "test"
				repository  = github_repository.test.id
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~ALL"]
						exclude = []
					}
				}

				rules {
					branch_name_pattern {
						name     = "test"
						negate   = false
						operator = "starts_with"
						pattern  = "test"
					}
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test", "name",
				"test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_ruleset.test", "enforcement",
				"active",
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

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Updates a ruleset name without error", func(t *testing.T) {

		repoName := fmt.Sprintf(`tf-acc-test-rename-%[1]s`, randomID)
		oldRSName := fmt.Sprintf(`ruleset-%[1]s`, randomID)
		newRSName := fmt.Sprintf(`%[1]s-renamed`, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name         = "%[1]s"
			  description  = "Terraform acceptance tests %[2]s"
			  vulnerability_alerts = true
			}

			resource "github_repository_ruleset" "test" {
				name        = "%[3]s"
				repository  = github_repository.test.id
				target      = "branch"
				enforcement = "active"

				rules {
					creation = true
				}
			}
		`, repoName, randomID, oldRSName)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_ruleset.test", "name",
					oldRSName,
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_ruleset.test", "name",
					newRSName,
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						// Rename the ruleset to something else
						Config: strings.Replace(
							config,
							oldRSName,
							newRSName, 1),
						Check: checks["after"],
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

	t.Run("Imports rulesets without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name         = "tf-acc-test-import-%[1]s"
			  description  = "Terraform acceptance tests %[1]s"
			  auto_init 	 = false
			  vulnerability_alerts = true
			}

			resource "github_repository_environment" "example" {
				environment  = "test"
				repository   = github_repository.test.name
			}

			resource "github_repository_ruleset" "test" {
				name        = "test"
				repository  = github_repository.test.id
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~ALL"]
						exclude = []
					}
				}

				rules {
					creation = true

					update = true

					deletion                = true
					required_linear_history = true

					required_deployments {
						required_deployment_environments = ["test"]
					}

					required_signatures = false

					pull_request {
						required_approving_review_count   = 2
						required_review_thread_resolution = true
						require_code_owner_review         = true
						dismiss_stale_reviews_on_push     = true
						require_last_push_approval        = true
					}

					required_status_checks {

						required_check {
							context = "ci"
						}

						strict_required_status_checks_policy = true
						do_not_enforce_on_create             = true
					}

					non_fast_forward = true
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_repository_ruleset.test", "name"),
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
					{
						ResourceName:      "github_repository_ruleset.test",
						ImportState:       true,
						ImportStateVerify: true,
						ImportStateIdFunc: importRepositoryRulesetByResourcePaths(
							"github_repository.test", "github_repository_ruleset.test"),
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

func importRepositoryRulesetByResourcePaths(repoLogicalName, rulesetLogicalName string) resource.ImportStateIdFunc {
	// test importing using an ID of the form <repo-node-id>:<ruleset-id>
	return func(s *terraform.State) (string, error) {
		log.Printf("[DEBUG] Looking up tf state ")
		repo := s.RootModule().Resources[repoLogicalName]
		if repo == nil {
			return "", fmt.Errorf("Cannot find %s in terraform state", repoLogicalName)
		}
		repoID := repo.Primary.ID
		if repoID == "" {
			return "", fmt.Errorf("repository %s does not have an id in terraform state", repoLogicalName)
		}

		ruleset := s.RootModule().Resources[rulesetLogicalName]
		if ruleset == nil {
			return "", fmt.Errorf("Cannot find %s in terraform state", rulesetLogicalName)
		}
		rulesetID := ruleset.Primary.ID
		if rulesetID == "" {
			return "", fmt.Errorf("ruleset %s does not have an id in terraform state", rulesetLogicalName)
		}

		return fmt.Sprintf("%s:%s", repoID, rulesetID), nil
	}
}
