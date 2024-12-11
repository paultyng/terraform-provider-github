package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryWebhooksDataSource(t *testing.T) {
	t.Run("manages repository webhooks", func(t *testing.T) {
		repoName := fmt.Sprintf("tf-acc-test-webhooks-%s", acctest.RandString(5))

		config := fmt.Sprintf(`
		resource "github_repository" "test" {
			name      = "%s"
			auto_init = true
		}

		resource "github_repository_webhook" "test" {
			repository = github_repository.test.name

			configuration {
				url          = "https://google.de/webhook"
				content_type = "json"
				insecure_ssl = true
			}

			events = ["pull_request"]
		}

		data "github_repository_webhooks" "test" {
			repository = github_repository.test.name

			depends_on = [github_repository_webhook.test]
		}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_repository_webhooks.test", "webhooks.#"),
						resource.TestCheckResourceAttr("data.github_repository_webhooks.test", "webhooks.#", "1"),
						resource.TestCheckResourceAttrSet("data.github_repository_webhooks.test", "webhooks.0.id"),
						resource.TestCheckResourceAttr("data.github_repository_webhooks.test", "webhooks.0.name", "web"),
						resource.TestCheckResourceAttr("data.github_repository_webhooks.test", "webhooks.0.active", "true"),
						resource.TestCheckResourceAttrSet("data.github_repository_webhooks.test", "webhooks.0.url"),
					),
				},
			},
		})
	})
}
