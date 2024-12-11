package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationWebhooksDataSource(t *testing.T) {
	t.Run("manages organization webhooks", func(t *testing.T) {
		config := `
		resource "github_organization_webhook" "test" {
			configuration {
				url          = "https://google.de/webhook"
				content_type = "json"
				insecure_ssl = true
			}

			events = ["pull_request"]
		}

		data "github_organization_webhooks" "test" {
		  depends_on = [github_organization_webhook.test]
		}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_organization_webhooks.test", "webhooks.#"),
						resource.TestCheckResourceAttr("data.github_organization_webhooks.test", "webhooks.#", "1"),
						resource.TestCheckResourceAttrSet("data.github_organization_webhooks.test", "webhooks.0.id"),
						resource.TestCheckResourceAttr("data.github_organization_webhooks.test", "webhooks.0.name", "web"),
						resource.TestCheckResourceAttr("data.github_organization_webhooks.test", "webhooks.0.active", "true"),
						resource.TestCheckResourceAttrSet("data.github_organization_webhooks.test", "webhooks.0.url"),
					),
				},
			},
		})
	})
}
