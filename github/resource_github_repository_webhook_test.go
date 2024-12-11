package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryWebhook(t *testing.T) {
	t.Run("creates repository webhooks without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name         = "test-%[1]s"
			  description  = "Terraform acceptance tests"
			}

			resource "github_repository_webhook" "test" {
			  depends_on = ["github_repository.test"]
			  repository = "test-%[1]s"

			  configuration {
			    url          = "https://google.de/webhook"
			    content_type = "json"
			    insecure_ssl = true
			  }

			  events = ["pull_request"]
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "active", "true",
			),
			resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "events.#", "1",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("imports repository webhooks without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name         = "test-%[1]s"
				description  = "Terraform acceptance tests"
			}

			resource "github_repository_webhook" "test" {
				depends_on = ["github_repository.test"]
				repository = "test-%[1]s"
				configuration {
					url          = "https://google.de/webhook"
					content_type = "json"
					insecure_ssl = true
				}
				events = ["pull_request"]
			}
			`, randomID)

		check := resource.ComposeTestCheckFunc()

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:        "github_repository_webhook.test",
					ImportState:         true,
					ImportStateVerify:   true,
					ImportStateIdPrefix: fmt.Sprintf("test-%s/", randomID),
				},
			},
		})
	})

	t.Run("updates repository webhooks without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		configs := map[string]string{
			"before": fmt.Sprintf(`
				resource "github_repository" "test" {
				  name         = "test-%[1]s"
				  description  = "Terraform acceptance tests"
				}

				resource "github_repository_webhook" "test" {
				  depends_on = ["github_repository.test"]
				  repository = "test-%[1]s"

				  configuration {
				    url          = "https://google.de/webhook"
				    content_type = "json"
				    insecure_ssl = true
				  }

				  events = ["pull_request"]
				}
			`, randomID),
			"after": fmt.Sprintf(`
				resource "github_repository" "test" {
				  name         = "test-%[1]s"
				  description  = "Terraform acceptance tests"
				}

				resource "github_repository_webhook" "test" {
				  depends_on = ["github_repository.test"]
				  repository = "test-%[1]s"

				  configuration {
				    secret       = "secret"
				    url          = "https://google.de/webhook"
				    content_type = "json"
				    insecure_ssl = true
				  }

				  events = ["pull_request"]
				}
			`, randomID),
		}

		checks := map[string]resource.TestCheckFunc{
			"before": resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "events.#", "1",
			),
			"after": resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "configuration.0.secret", "secret",
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
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
}
