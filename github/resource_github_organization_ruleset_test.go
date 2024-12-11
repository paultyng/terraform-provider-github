package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestGithubOrganizationRulesets(t *testing.T) {
	t.Run("Creates and updates organization rulesets without errors", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-%s"
				target      = "branch"
				enforcement = "active"

				conditions {
				  repository_name {
						include = ["~ALL"]
						exclude = []
					}

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
					}

					required_workflows {
						required_workflow {
							path          = "path/to/workflow.yaml"
							repository_id = 1234
						}
					}

					required_code_scanning {
					  required_code_scanning_tool {
						alerts_threshold = "errors"
						security_alerts_threshold = "high_or_higher"
						tool = "CodeQL"
					  }
					}

					branch_name_pattern {
						name     = "test"
						negate   = false
						operator = "starts_with"
						pattern  = "test"
					}

					non_fast_forward = true
				}
			}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("github_organization_ruleset.test", "name", "test"),
						resource.TestCheckResourceAttr("github_organization_ruleset.test", "enforcement", "active"),
					),
				},
			},
		})
	})

	t.Run("Updates a ruleset name without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		name := fmt.Sprintf("test-acc-ruleset-%s", randomID)
		nameUpdated := fmt.Sprintf("test-acc-ruleset-updated-%s", randomID)

		config := fmt.Sprintf(`
		resource "github_organization_ruleset" "test" {
			name        = "%s"
			target      = "branch"
			enforcement = "active"

			conditions {
				repository_name {
					include = ["~ALL"]
					exclude = []
				}

				ref_name {
					include = ["~ALL"]
					exclude = []
				}
			}

			rules {
				creation = true
			}
		}
		`, name)

		configUpdated := fmt.Sprintf(`
		resource "github_organization_ruleset" "test" {
			name        = "%s"
			target      = "branch"
			enforcement = "active"

			conditions {
				repository_name {
					include = ["~ALL"]
					exclude = []
				}

				ref_name {
					include = ["~ALL"]
					exclude = []
				}
			}

			rules {
				creation = true
			}
		}
		`, nameUpdated)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("github_organization_ruleset.test", "name", name),
					),
				},
				{
					Config: configUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("github_organization_ruleset.test", "name", nameUpdated),
					),
				},
			},
		})
	})

	t.Run("imports rulesets without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_organization_ruleset" "test" {
				name        = "test-%s"
				target      = "branch"
				enforcement = "active"

				conditions {
				  repository_name {
						include = ["~ALL"]
						exclude = []
					}

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
					}

					branch_name_pattern {
						name     = "test"
						negate   = false
						operator = "starts_with"
						pattern  = "test"
					}

					non_fast_forward = true
				}
			}


		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					ResourceName:      "github_organization_ruleset.test",
					ImportState:       true,
					ImportStateVerify: true,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_organization_ruleset.test", "name"),
					),
				},
			},
		})
	})
}
