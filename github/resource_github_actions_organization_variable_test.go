package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsOrganizationVariable(t *testing.T) {
	t.Run("creates and updates a private organization variable without error", func(t *testing.T) {
		value := "my_variable_value"
		updatedValue := "my_updated_variable_value"

		config := fmt.Sprintf(`
			resource "github_actions_organization_variable" "variable" {
			  variable_name    = "test_variable"
			  value  		   = "%s"
			  visibility       = "private"
			}
			`, value)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_organization_variable.variable", "value",
					value,
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_variable.variable", "visibility",
					"private",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_variable.variable", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_variable.variable", "updated_at",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_organization_variable.variable", "value",
					updatedValue,
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_variable.variable", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_variable.variable", "updated_at",
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks["before"],
				},
				{
					Config: strings.Replace(config,
						value,
						updatedValue, 1),
					Check: checks["after"],
				},
			},
		})
	})

	t.Run("creates an organization variable scoped to a repo without error", func(t *testing.T) {
		value := "my_variable_value"
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
			}

			resource "github_actions_organization_variable" "variable" {
			  variable_name    = "test_variable"
			  value  		   = "%s"
			  visibility       = "selected"
			  selected_repository_ids = [github_repository.test.repo_id]
			}
			`, randomID, value)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_organization_variable.variable", "value",
					value,
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_variable.variable", "visibility",
					"selected",
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_variable.variable", "selected_repository_ids.#",
					"1",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_variable.variable", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_variable.variable", "updated_at",
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks["before"],
				},
			},
		})
	})

	t.Run("deletes organization variables without error", func(t *testing.T) {
		config := `
				resource "github_actions_organization_variable" "variable" {
				variable_name    = "test_variable"
				value  = "my_variable_value"
				visibility       = "private"
				}
			`

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:  config,
					Destroy: true,
				},
			},
		})
	})

	t.Run("imports an organization variable without error", func(t *testing.T) {
		value := "my_variable_value"
		varName := "test_variable"

		config := fmt.Sprintf(`
			resource "github_actions_organization_variable" "variable" {
			  variable_name    = "%s"
			  value  		   = "%s"
			  visibility       = "private"
			}
			`, varName, value)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					ResourceName:      "github_actions_organization_variable.variable",
					ImportStateId:     varName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
