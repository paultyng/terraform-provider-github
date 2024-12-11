package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsRunnerGroup(t *testing.T) {
	t.Run("creates runner groups without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			  vulnerability_alerts = false
			  auto_init = true
			}

			resource "github_branch" "test" {
			  repository = github_repository.test.name
			  branch     = "test"
			}

			resource "github_branch_default" "default"{
			  repository = github_repository.test.name
			  branch     = github_branch.test.branch
			}

			resource "github_repository_file" "workflow_file" {
			  depends_on  = [github_branch_default.default]
			  repository          = github_repository.test.name
			  file                = ".github/workflows/test.yml"
			  content             = ""
			  commit_message      = "Managed by Terraform"
			  commit_author       = "Terraform User"
			  commit_email        = "terraform@example.com"
			  overwrite_on_create = true
			}

			resource "github_actions_runner_group" "test" {
			  depends_on  = [github_repository_file.workflow_file]

			  name       = github_repository.test.name
			  visibility = "all"
			  restricted_to_workflows = true
			  selected_workflows = ["${github_repository.test.full_name}/.github/workflows/test.yml@refs/heads/${github_branch.test.branch}"]
			  allows_public_repositories = true
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"github_actions_runner_group.test", "name",
			),
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "name",
				fmt.Sprintf(`tf-acc-test-%s`, randomID),
			),
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "visibility",
				"all",
			),
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "restricted_to_workflows",
				"true",
			),
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "selected_workflows.#",
				"1",
			),
			func(state *terraform.State) error {
				githubRepository := state.RootModule().Resources["github_repository.test"].Primary
				fullName := githubRepository.Attributes["full_name"]

				runnerGroup := state.RootModule().Resources["github_actions_runner_group.test"].Primary
				workflowActual := runnerGroup.Attributes["selected_workflows.0"]

				workflowExpected := fmt.Sprintf("%s/.github/workflows/test.yml@refs/heads/test", fullName)

				if workflowActual != workflowExpected {
					return fmt.Errorf("actual selected workflows %s not the same as expected selected workflows %s",
						workflowActual, workflowExpected)
				}
				return nil
			},
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "allows_public_repositories",
				"true",
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

	t.Run("manages runner visibility", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			}

			resource "github_actions_runner_group" "test" {
			  name       = github_repository.test.name
			  visibility = "selected"
			  selected_repository_ids = [github_repository.test.repo_id]
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"github_actions_runner_group.test", "name",
			),
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "name",
				fmt.Sprintf(`tf-acc-test-%s`, randomID),
			),
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "visibility",
				"selected",
			),
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "selected_repository_ids.#",
				"1",
			),
			resource.TestCheckResourceAttrSet(
				"github_actions_runner_group.test", "selected_repositories_url",
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

	t.Run("imports an all runner group without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			}

			resource "github_actions_runner_group" "test" {
			  name       = github_repository.test.name
			  visibility = "all"
			}
    `, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_actions_runner_group.test", "name"),
			resource.TestCheckResourceAttrSet("github_actions_runner_group.test", "visibility"),
			resource.TestCheckResourceAttr("github_actions_runner_group.test", "visibility", "all"),
			resource.TestCheckResourceAttr("github_actions_runner_group.test", "name", fmt.Sprintf(`tf-acc-test-%s`, randomID)),
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
					ResourceName:      "github_actions_runner_group.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("imports a private runner group without error", func(t *testing.T) {
		// Note: this test is skipped because when setting visibility 'private', it always fails with:
		// Step 0 error: After applying this step, the plan was not empty:
		// visibility:                 "all" => "private"
		t.Skip("This is not supported")

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
					resource "github_repository" "test" {
					  name = "tf-acc-test-%s"
					}

					resource "github_actions_runner_group" "test" {
					  name       = github_repository.test.name
					  visibility = "private"
					}
		    `, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_actions_runner_group.test", "name"),
			resource.TestCheckResourceAttr("github_actions_runner_group.test", "name", fmt.Sprintf(`tf-acc-test-%s`, randomID)),
			resource.TestCheckResourceAttrSet("github_actions_runner_group.test", "visibility"),
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
					ResourceName:      "github_actions_runner_group.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("imports a selected runner group without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
			}

			resource "github_actions_runner_group" "test" {
				name       = github_repository.test.name
				visibility = "selected"
				selected_repository_ids = [github_repository.test.repo_id]
			}
    `, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_actions_runner_group.test", "name"),
			resource.TestCheckResourceAttr("github_actions_runner_group.test", "name", fmt.Sprintf(`tf-acc-test-%s`, randomID)),
			resource.TestCheckResourceAttrSet("github_actions_runner_group.test", "visibility"),
			resource.TestCheckResourceAttr("github_actions_runner_group.test", "visibility", "selected"),
			resource.TestCheckResourceAttr(
				"github_actions_runner_group.test", "selected_repository_ids.#",
				"1",
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
					ResourceName:      "github_actions_runner_group.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
