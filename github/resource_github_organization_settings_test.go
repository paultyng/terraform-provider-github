package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationSettings(t *testing.T) {
	t.Run("creates organization settings without error", func(t *testing.T) {
		config := `
		resource "github_organization_settings" "test" {
			billing_email = "test@example.com"
			company = "Test Company"
			blog = "https://example.com"
			email = "test@example.com"
			twitter_username = "Test"
			location = "Test Location"
			name = "Test Name"
			description = "Test Description"
			has_organization_projects = true
			has_repository_projects = true
			default_repository_permission = "read"
			members_can_create_repositories = true
			members_can_create_public_repositories = true
			members_can_create_private_repositories = true
			members_can_create_internal_repositories = false
			members_can_create_pages = true
			members_can_create_public_pages = true
			members_can_create_private_pages = true
			members_can_fork_private_repositories = true
			web_commit_signoff_required = true
			advanced_security_enabled_for_new_repositories = false
			  dependabot_alerts_enabled_for_new_repositories=  false
			dependabot_security_updates_enabled_for_new_repositories = false
			dependency_graph_enabled_for_new_repositories = false
			secret_scanning_enabled_for_new_repositories = false
			secret_scanning_push_protection_enabled_for_new_repositories = false
		  }`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_organization_settings.test",
				"billing_email", "test@example.com",
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

	t.Run("updates organization settings without error", func(t *testing.T) {
		billingEmail := "test1@example.com"
		company := "Test Company"
		blog := "https://test.com"
		updatedBillingEmail := "test2@example.com"
		updatedCompany := "Test Company 2"
		updatedBlog := "https://test2.com"

		configs := map[string]string{
			"before": fmt.Sprintf(`
			resource "github_organization_settings" "test" {
				billing_email = "%s"
				company = "%s"
				blog = "%s"
				}`, billingEmail, company, blog),

			"after": fmt.Sprintf(`
			resource "github_organization_settings" "test" {
				billing_email = "%s"
				company = "%s"
				blog = "%s"
				}`, updatedBillingEmail, updatedCompany, updatedBlog),
		}
		checks := map[string]resource.TestCheckFunc{
			"before": resource.TestCheckResourceAttr(
				"github_organization_settings.test",
				"billing_email", billingEmail,
			),
			"after": resource.TestCheckResourceAttr(
				"github_organization_settings.test",
				"billing_email", updatedBillingEmail,
			),
		}
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: configs["before"],
					Check:  checks["before"],
				},
				{
					Config: configs["after"],
					Check:  checks["after"],
				},
			},
		})
	})

	t.Run("imports organization settings without error", func(t *testing.T) {
		billingEmail := "test@example.com"
		company := "Test Company"
		blog := "https://example.com"

		config := fmt.Sprintf(`
		resource "github_organization_settings" "test" {
			billing_email = "%s"
			company = "%s"
			blog = "%s"
			}`, billingEmail, company, blog)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_organization_settings.test",
				"billing_email", billingEmail,
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
				{
					ResourceName:      "github_organization_settings.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
