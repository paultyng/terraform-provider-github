package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryDependabotSecurityUpdates(t *testing.T) {
	t.Run("enables automated security fixes without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		enabled := "enabled = false"
		updatedEnabled := "enabled = true"
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				visibility = "private"
			  	auto_init = true
				vulnerability_alerts   = true
			}


			resource "github_repository_dependabot_security_updates" "test" {
			  repository  = github_repository.test.id
			  %s
			}
		`, randomID, enabled)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_dependabot_security_updates.test", "enabled",
					"false",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_dependabot_security_updates.test", "enabled",
					"true",
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
						enabled,
						updatedEnabled, 1),
					Check: checks["after"],
				},
			},
		})
	})

	t.Run("disables automated security fixes without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		enabled := "enabled = true"
		updatedEnabled := "enabled = false"

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				visibility = "private"
			  	auto_init = true
				vulnerability_alerts   = true
			}


			resource "github_repository_dependabot_security_updates" "test" {
			  repository  = github_repository.test.id
			  %s
			}
		`, randomID, enabled)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_dependabot_security_updates.test", "enabled",
					"true",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_dependabot_security_updates.test", "enabled",
					"false",
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
						enabled,
						updatedEnabled, 1),
					Check: checks["after"],
				},
			},
		})
	})

	t.Run("imports automated security fixes without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			  vulnerability_alerts   = true
			}

			resource "github_repository_dependabot_security_updates" "test" {
			  repository  = github_repository.test.id
			  enabled = false
			}
    `, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_repository_dependabot_security_updates.test", "repository"),
			resource.TestCheckResourceAttrSet("github_repository_dependabot_security_updates.test", "enabled"),
			resource.TestCheckResourceAttr("github_repository_dependabot_security_updates.test", "enabled", "false"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:      "github_repository_dependabot_security_updates.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
