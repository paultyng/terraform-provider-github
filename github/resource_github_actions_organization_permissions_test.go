package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsOrganizationPermissions(t *testing.T) {

	t.Run("test setting of basic actions organization permissions", func(t *testing.T) {

		allowedActions := "local_only"
		enabledRepositories := "all"

		config := fmt.Sprintf(`
			resource "github_actions_organization_permissions" "test" {
				allowed_actions = "%s"
				enabled_repositories = "%s"
			}
		`, allowedActions, enabledRepositories)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "enabled_repositories", enabledRepositories,
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

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("imports entire set of github action organization permissions without error", func(t *testing.T) {

		allowedActions := "selected"
		enabledRepositories := "selected"
		githubOwnedAllowed := true
		verifiedAllowed := true
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-topic-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics			= ["terraform", "testing"]
			}

			resource "github_actions_organization_permissions" "test" {
				allowed_actions = "%s"
				enabled_repositories = "%s"
				allowed_actions_config {
					github_owned_allowed = %t
					patterns_allowed     = ["actions/cache@*", "actions/checkout@*"]
					verified_allowed     = %t
				}
				enabled_repositories_config {
					repository_ids       = [github_repository.test.repo_id]
				}
			}
		`, randomID, allowedActions, enabledRepositories, githubOwnedAllowed, verifiedAllowed)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "enabled_repositories", enabledRepositories,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "allowed_actions_config.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "enabled_repositories_config.#", "1",
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
					{
						ResourceName:      "github_actions_organization_permissions.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		}

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("test setting of organization allowed actions", func(t *testing.T) {

		allowedActions := "selected"
		enabledRepositories := "all"
		githubOwnedAllowed := true
		verifiedAllowed := true

		config := fmt.Sprintf(`

			resource "github_actions_organization_permissions" "test" {
				allowed_actions = "%s"
				enabled_repositories = "%s"
				allowed_actions_config {
					github_owned_allowed = %t
					patterns_allowed     = ["actions/cache@*", "actions/checkout@*"]
					verified_allowed     = %t
				}
			}
		`, allowedActions, enabledRepositories, githubOwnedAllowed, verifiedAllowed)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "enabled_repositories", enabledRepositories,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "allowed_actions_config.#", "1",
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

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("test not setting of organization allowed actions without error", func(t *testing.T) {

		allowedActions := "selected"
		enabledRepositories := "all"

		config := fmt.Sprintf(`

			resource "github_actions_organization_permissions" "test" {
				allowed_actions = "%s"
				enabled_repositories = "%s"
			}
		`, allowedActions, enabledRepositories)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "enabled_repositories", enabledRepositories,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "allowed_actions_config.#", "0",
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

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("test setting of organization enabled repositories", func(t *testing.T) {

		allowedActions := "all"
		enabledRepositories := "selected"
		githubOwnedAllowed := true
		verifiedAllowed := true
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		randomID2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-topic-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics			= ["terraform", "testing"]
			}

			resource "github_repository" "test2" {
				name        = "tf-acc-test-topic-%[2]s"
				description = "Terraform acceptance tests %[2]s"
				topics			= ["terraform", "testing"]
			}

			resource "github_actions_organization_permissions" "test" {
				allowed_actions = "%s"
				enabled_repositories = "%s"
				enabled_repositories_config {
					repository_ids       = [github_repository.test.repo_id, github_repository.test2.repo_id]
				}
			}
		`, randomID, randomID2, allowedActions, enabledRepositories, githubOwnedAllowed, verifiedAllowed)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "enabled_repositories", enabledRepositories,
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "enabled_repositories_config.#", "1",
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

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("test disabling of actions at organization level", func(t *testing.T) {

		enabledRepositories := "none"

		config := fmt.Sprintf(`
			resource "github_actions_organization_permissions" "test" {
				enabled_repositories = "%s"
			}
		`, enabledRepositories)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_permissions.test", "enabled_repositories", enabledRepositories,
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

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

}
