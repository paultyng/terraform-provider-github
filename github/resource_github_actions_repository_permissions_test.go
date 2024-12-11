package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsRepositoryPermissions(t *testing.T) {
	t.Run("test setting of basic actions repository permissions", func(t *testing.T) {
		allowedActions := "local_only"
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-topic-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics		= ["terraform", "testing"]
			}

			resource "github_actions_repository_permissions" "test" {
				allowed_actions = "%s"
				repository = github_repository.test.name
			}
		`, randomID, allowedActions)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "allowed_actions", allowedActions,
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

	t.Run("imports entire set of github action repository permissions without error", func(t *testing.T) {
		allowedActions := "selected"
		githubOwnedAllowed := true
		verifiedAllowed := true
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-topic-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics		= ["terraform", "testing"]
			}

			resource "github_actions_repository_permissions" "test" {
				allowed_actions = "%s"
				allowed_actions_config {
					github_owned_allowed = %t
					patterns_allowed     = ["actions/cache@*", "actions/checkout@*"]
					verified_allowed     = %t
				}
				repository = github_repository.test.name
			}
		`, randomID, allowedActions, githubOwnedAllowed, verifiedAllowed)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "allowed_actions_config.#", "1",
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
				{
					ResourceName:      "github_actions_repository_permissions.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("test setting of repository allowed actions", func(t *testing.T) {
		allowedActions := "selected"
		githubOwnedAllowed := true
		verifiedAllowed := true
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-topic-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics		= ["terraform", "testing"]
			}

			resource "github_actions_repository_permissions" "test" {
				allowed_actions = "%s"
				allowed_actions_config {
					github_owned_allowed = %t
					patterns_allowed     = ["actions/cache@*", "actions/checkout@*"]
					verified_allowed     = %t
				}
				repository = github_repository.test.name
			}
		`, randomID, allowedActions, githubOwnedAllowed, verifiedAllowed)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "allowed_actions_config.#", "1",
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

	t.Run("test not setting of repository allowed actions without error", func(t *testing.T) {
		allowedActions := "selected"
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-topic-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics		= ["terraform", "testing"]
			}

			resource "github_actions_repository_permissions" "test" {
				allowed_actions = "%s"
				repository = github_repository.test.name
			}
		`, randomID, allowedActions)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "allowed_actions", allowedActions,
			),
			// Even if we do not set the allowed_actions_config,
			// it will be set to the organization level settings
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "allowed_actions_config.#", "0",
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

	t.Run("test disabling actions on a repository", func(t *testing.T) {
		actionsEnabled := false
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-topic-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics		= ["terraform", "testing"]
			}

			resource "github_actions_repository_permissions" "test" {
				enabled = %t
				repository = github_repository.test.name
			}
		`, randomID, actionsEnabled)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "enabled", "false",
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

	// https://github.com/integrations/terraform-provider-github/issues/2182
	t.Run("test load with disabled actions", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			locals {
				actions_enabled = false
			}

			resource "github_repository" "test" {
				name        = "tf-acc-test-actions-permissions-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics		= ["terraform", "testing"]
			}

			resource "github_actions_repository_permissions" "test" {
				repository = github_repository.test.name
				enabled         = local.actions_enabled
				allowed_actions = local.actions_enabled ? "all" : null
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "enabled", "false",
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_permissions.test", "allowed_actions.#", "0",
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
