package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsEnvironmentVariable(t *testing.T) {
	t.Run("creates and updates environment variables without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		value := "my_variable_value"
		updatedValue := "my_updated_variable_value"

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			}

			resource "github_repository_environment" "test" {
			  repository       = github_repository.test.name
			  environment      = "environment / test"
			}

			resource "github_actions_environment_variable" "variable" {
			  repository       = github_repository.test.name
			  environment      = github_repository_environment.test.environment
			  variable_name    = "test_variable"
			  value  = "%s"
			}
			`, randomID, value)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_environment_variable.variable", "value",
					value,
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_environment_variable.variable", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_environment_variable.variable", "updated_at",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_environment_variable.variable", "value",
					updatedValue,
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_environment_variable.variable", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_environment_variable.variable", "updated_at",
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

	t.Run("deletes environment variables without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
				resource "github_repository" "test" {
					name = "tf-acc-test-%s"
				}

				resource "github_repository_environment" "test" {
					repository       = github_repository.test.name
					environment      = "environment / test"
				}

				resource "github_actions_environment_variable" "variable" {
					repository 		= github_repository.test.name
					environment     = github_repository_environment.test.environment
					variable_name	= "test_variable"
					value 			= "my_variable_value"
				}
			`, randomID)

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

	t.Run("imports environment variables without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		value := "my_variable_value"
		envName := "environment / test"
		varName := "test_variable"

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			}

			resource "github_repository_environment" "test" {
			  repository       = github_repository.test.name
			  environment      = "%s"
			}

			resource "github_actions_environment_variable" "variable" {
			  repository       = github_repository.test.name
			  environment      = github_repository_environment.test.environment
			  variable_name    = "%s"
			  value  = "%s"
			}
			`, randomID, envName, varName, value)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					ResourceName:      "github_actions_environment_variable.variable",
					ImportStateId:     fmt.Sprintf(`tf-acc-test-%s:%s:%s`, randomID, envName, varName),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
