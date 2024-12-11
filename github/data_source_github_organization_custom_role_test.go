package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationCustomRoleDataSource(t *testing.T) {
	t.Run("queries a custom repo role", func(t *testing.T) {
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

		config2 := config + `
			data "github_organization_custom_role" "test" {
				name = github_organization_custom_role.test.name
			}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"data.github_organization_custom_role.test", "name",
			),
			resource.TestCheckResourceAttr(
				"data.github_organization_custom_role.test", "name",
				fmt.Sprintf(`tf-acc-test-%s`, randomID),
			),
			resource.TestCheckResourceAttr(
				"data.github_organization_custom_role.test", "description",
				"Test role description",
			),
			resource.TestCheckResourceAttr(
				"data.github_organization_custom_role.test", "base_role",
				"read",
			),
			resource.TestCheckResourceAttr(
				"data.github_organization_custom_role.test", "permissions.#",
				"2",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  resource.ComposeTestCheckFunc(),
				},
				{
					Config: config2,
					Check:  check,
				},
			},
		})
	})
}
