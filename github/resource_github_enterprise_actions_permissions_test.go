package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsEnterprisePermissions(t *testing.T) {
	t.Run("test setting of basic actions enterprise permissions", func(t *testing.T) {
		allowedActions := "local_only"
		enabledOrganizations := "all"

		config := fmt.Sprintf(`
			resource "github_enterprise_actions_permissions" "test" {
				enterprise_slug = "%s"
				allowed_actions = "%s"
				enabled_organizations = "%s"
			}
		`, testAccConf.enterpriseSlug, allowedActions, enabledOrganizations)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "enabled_organizations", enabledOrganizations,
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("imports entire set of github action enterprise permissions without error", func(t *testing.T) {
		allowedActions := "selected"
		enabledOrganizations := "selected"
		githubOwnedAllowed := true
		verifiedAllowed := true
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-displayname%s", randomID)

		displayName := fmt.Sprintf("Tf Acc Test displayname %s", randomID)

		desc := "Initial org description"

		config := fmt.Sprintf(`
			data "github_user" "current" {
				username = ""
			}

			resource "github_enterprise_organization" "org" {
				enterprise_slug = "%s"
				name            = "%s"
				display_name    = "%s"
				description     = "%s"
				billing_email   = data.github_user.current.email
				admin_logins    = [
					data.github_user.current.login
				]
			}

			resource "github_enterprise_actions_permissions" "test" {
				enterprise_slug = "%s"
				allowed_actions = "%s"
				enabled_organizations = "%s"
				allowed_actions_config {
					github_owned_allowed = %t
					patterns_allowed     = ["actions/cache@*", "actions/checkout@*"]
					verified_allowed     = %t
				}
				enabled_organizations_config {
					organization_ids       = [github_enterprise_organization.org.id]
				}
			}
		`, testAccConf.enterpriseSlug, orgName, displayName, desc, testAccConf.enterpriseSlug, allowedActions, enabledOrganizations, githubOwnedAllowed, verifiedAllowed)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "enabled_organizations", enabledOrganizations,
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "allowed_actions_config.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "enabled_organizations_config.#", "1",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:      "github_enterprise_actions_permissions.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("test setting of enterprise allowed actions", func(t *testing.T) {
		allowedActions := "selected"
		enabledOrganizations := "all"
		githubOwnedAllowed := true
		verifiedAllowed := true

		config := fmt.Sprintf(`
			resource "github_enterprise_actions_permissions" "test" {
				enterprise_slug = "%s"
				allowed_actions = "%s"
				enabled_organizations = "%s"
				allowed_actions_config {
					github_owned_allowed = %t
					patterns_allowed     = ["actions/cache@*", "actions/checkout@*"]
					verified_allowed     = %t
				}
			}
		`, testAccConf.enterpriseSlug, allowedActions, enabledOrganizations, githubOwnedAllowed, verifiedAllowed)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "enabled_organizations", enabledOrganizations,
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "allowed_actions_config.#", "1",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("test setting of enterprise enabled organizations", func(t *testing.T) {
		allowedActions := "all"
		enabledOrganizations := "selected"
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		randomID2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-displayname%s", randomID)
		orgName2 := fmt.Sprintf("tf-acc-test-displayname%s", randomID2)

		displayName := fmt.Sprintf("Tf Acc Test displayname %s", randomID)
		displayName2 := fmt.Sprintf("Tf Acc Test displayname %s", randomID2)

		desc := fmt.Sprintf("Initial org description %s", randomID)
		desc2 := fmt.Sprintf("Initial org description %s", randomID2)

		config := fmt.Sprintf(`
			data "github_user" "current" {
				username = ""
			}
			resource "github_enterprise_organization" "org" {
				enterprise_slug = "%s"
				name            = "%s"
				display_name    = "%s"
				description     = "%s"
				billing_email   = data.github_user.current.email
				admin_logins    = [
					data.github_user.current.login
				]
			}
			resource "github_enterprise_organization" "org2" {
				enterprise_slug = "%s"
				name            = "%s"
				display_name    = "%s"
				description     = "%s"
				billing_email   = data.github_user.current.email
				admin_logins    = [
					data.github_user.current.login
				]
			}
			resource "github_enterprise_actions_permissions" "test" {
				enterprise_slug = "%s"
				allowed_actions = "%s"
				enabled_organizations = "%s"
				enabled_organizations_config {
					organization_ids       = [github_enterprise_organization.org.id, github_enterprise_organization.org2.id]
				}
			}
		`, testAccConf.enterpriseSlug, orgName, displayName, desc, testAccConf.enterpriseSlug, orgName2, displayName2, desc2, testAccConf.enterpriseSlug, allowedActions, enabledOrganizations)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "allowed_actions", allowedActions,
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "enabled_organizations", enabledOrganizations,
			),
			resource.TestCheckResourceAttr(
				"github_enterprise_actions_permissions.test", "enabled_organizations_config.#", "1",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
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
