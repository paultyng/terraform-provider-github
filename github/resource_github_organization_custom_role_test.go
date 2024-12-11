package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationCustomRole(t *testing.T) {
	t.Run("creates custom repo role without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
		resource "github_organization_custom_role" "test" {
			name        = "tf-acc-test-%s"
			description = "Test role description"
			base_role   = "read"
			permissions = [
				"reopen_issue",
				"reopen_pull_request",
			]
		}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_organization_custom_role.test", "name"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "name", fmt.Sprintf(`tf-acc-test-%s`, randomID)),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "description", "Test role description"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "base_role", "read"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "permissions.#", "2"),
					),
				},
			},
		})
	})

	// More tests can go here following the same format...
	t.Run("updates custom repo role without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
		resource "github_organization_custom_role" "test" {
			name        = "tf-acc-test-%s"
			description = "Updated test role description before"
			base_role   = "read"
			permissions = [
				"reopen_issue",
				"reopen_pull_request",
			]
		}
		`, randomID)

		configUpdated := fmt.Sprintf(`
		resource "github_organization_custom_role" "test" {
			name        = "tf-acc-test-rename-%s"
			description = "Updated test role description after"
			base_role   = "write"
			permissions = [
				"reopen_issue",
				"read_code_scanning",
				"reopen_pull_request",
			]
		}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_organization_custom_role.test", "name"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "name", fmt.Sprintf(`tf-acc-test-%s`, randomID)),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "description", "Updated test role description before"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "base_role", "read"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "permissions.#", "2"),
					),
				},
				{
					Config: configUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_organization_custom_role.test", "name"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "name", fmt.Sprintf(`tf-acc-test-rename-%s`, randomID)),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "description", "Updated test role description after"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "base_role", "write"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "permissions.#", "3"),
					),
				},
			},
		})
	})

	t.Run("imports custom repo role without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_organization_custom_role" "test" {
			  name        = "tf-acc-test-%s"
			  description = "Test role description"
			  base_role   = "read"
			  permissions = [
					"reopen_issue",
					"reopen_pull_request",
					"read_code_scanning"
				]
			}
    `, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_organization_custom_role.test", "name"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "name", fmt.Sprintf(`tf-acc-test-%s`, randomID)),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "description", "Test role description"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "base_role", "read"),
						resource.TestCheckResourceAttr("github_organization_custom_role.test", "permissions.#", "3"),
					),
				},
			},
		})
	})
}
