package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationRepositoriesDataSource(t *testing.T) {
	t.Run("manages repositories", func(t *testing.T) {
		config := `
			resource "github_repository" "test1" {
			  name       = "test1"
			  visibility = "private"
			}

			resource "github_repository" "test2" {
			  name       = "test2"
			  archived   = true
			  visibility = "public"
			  depends_on = [github_repository.test1]
			}
		 `

		config2 := config + `
			data "github_organization_repositories" "all" {}
		`

		const resourceName = "data.github_organization_repositories.all"
		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "webhooks.#", "2"),
			resource.TestCheckResourceAttr(resourceName, "webhooks.0.name", "test1"),
			resource.TestCheckResourceAttr(resourceName, "webhooks.0.archived", "false"),
			resource.TestCheckResourceAttr(resourceName, "webhooks.0.visibility", "private"),
			resource.TestCheckResourceAttrSet(resourceName, "webhooks.0.repo_id"),
			resource.TestCheckResourceAttrSet(resourceName, "webhooks.0.node_id"),
			resource.TestCheckResourceAttr(resourceName, "webhooks.1.name", "test2"),
			resource.TestCheckResourceAttr(resourceName, "webhooks.1.archived", "true"),
			resource.TestCheckResourceAttr(resourceName, "webhooks.1.visibility", "public"),
			resource.TestCheckResourceAttrSet(resourceName, "webhooks.1.repo_id"),
			resource.TestCheckResourceAttrSet(resourceName, "webhooks.1.node_id"),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
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
