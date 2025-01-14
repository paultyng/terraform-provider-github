package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubProjectCard(t *testing.T) {
	t.Skip("Skipping test as the GitHub REST API no longer supports classic projects")

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates a project card using a note", func(t *testing.T) {
		config := fmt.Sprintf(`

			resource "github_organization_project" "project" {
				name = "tf-acc-%s"
				body = "This is an organization project."
			}

			resource "github_project_column" "column" {
				project_id = github_organization_project.project.id
				name       = "Backlog"
			}

			resource "github_project_card" "card" {
				column_id = github_project_column.column.column_id
				note        = "## Unaccepted 👇"
			}

		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"github_project_card.card", "note",
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
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("creates a project card using an issue", func(t *testing.T) {
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
				has_projects = true
				has_issues   = true
			}

			resource "github_issue" "test" {
			  repository       = github_repository.test.id
			  title            = "Test issue title"
			  body             = "Test issue body"
			}

			resource "github_repository_project" "test" {
			  name            = "test"
			  repository      = github_repository.test.name
			  body            = "this is a test project"
			}

			resource "github_project_column" "test" {
				project_id = github_repository_project.test.id
				name       = "Backlog"
			}

			resource "github_project_card" "test" {
				column_id    = github_project_column.test.column_id
				content_id   = github_issue.test.issue_id
				content_type = "Issue"
			}

		`, randomID)

		check := resource.ComposeTestCheckFunc(
			func(state *terraform.State) error {
				issue := state.RootModule().Resources["github_issue.test"].Primary
				card := state.RootModule().Resources["github_project_card.test"].Primary

				issueID := issue.Attributes["issue_id"]
				cardID := card.Attributes["content_id"]
				if cardID != issueID {
					return fmt.Errorf("card content_id %s not the same as issue id %s",
						cardID, issueID)
				}
				return nil
			},
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
}
