package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryTopics(t *testing.T) {
	t.Run("create repository topics and import them", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_repository_topics" "test" {
				repository    = github_repository.test.name
				topics        = ["test", "test-2"]
			}
		`, randomID)

		const resourceName = "github_repository_topics.test"

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "topics.#", "2"),
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
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("create repository topics and update them", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		configBefore := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_repository_topics" "test" {
				repository    = github_repository.test.name
				topics        = ["test", "test-2"]
			}
		`, randomID)

		configAfter := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_repository_topics" "test" {
				repository    = github_repository.test.name
				topics        = ["test", "test-2", "extra-topic"]
			}
		`, randomID)

		const resourceName = "github_repository_topics.test"

		checkBefore := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "topics.#", "2"),
		)
		checkAfter := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "topics.#", "3"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: configBefore,
					Check:  checkBefore,
				},
				{
					Config: configAfter,
					Check:  checkAfter,
				},
			},
		})
	})
}
