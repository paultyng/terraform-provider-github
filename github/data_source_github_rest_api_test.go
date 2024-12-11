package github

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRestApiDataSource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("queries an existing branch without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%[1]s"
				auto_init = true
			}

			data "github_rest_api" "test" {
				endpoint = "repos/${github_repository.test.full_name}/git/refs/heads/${github_repository.test.default_branch}"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestMatchResourceAttr(
				"data.github_rest_api.test", "code", regexp.MustCompile("200"),
			),
			resource.TestMatchResourceAttr(
				"data.github_rest_api.test", "status", regexp.MustCompile("200 OK"),
			),
			resource.TestMatchResourceAttr("data.github_rest_api.test", "body", regexp.MustCompile(".*refs/heads/.*")),
			resource.TestCheckResourceAttrSet("data.github_rest_api.test", "headers"),
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

	t.Run("queries a collection without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%[1]s"
				auto_init = true
			}

			data "github_rest_api" "test" {
				endpoint = "repos/${github_repository.test.full_name}/git/refs/heads/"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestMatchResourceAttr("data.github_rest_api.test", "body", regexp.MustCompile(`\[.*refs/heads/.*\]`)),
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

	t.Run("queries an invalid branch without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%[1]s"
				auto_init = true
			}

			data "github_rest_api" "test" {
				endpoint = "repos/${github_repository.test.full_name}/git/refs/heads/xxxxxx"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestMatchResourceAttr(
				"data.github_rest_api.test", "code", regexp.MustCompile("404"),
			),
			resource.TestMatchResourceAttr(
				"data.github_rest_api.test", "status", regexp.MustCompile("404 Not Found"),
			),
			resource.TestCheckResourceAttrSet("data.github_rest_api.test", "body"),
			resource.TestCheckResourceAttrSet("data.github_rest_api.test", "headers"),
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

	t.Run("fails for invalid endpoint", func(t *testing.T) {
		// 4096 characters is the maximum length for a URL
		endpoint := strings.Repeat("x", 4096)
		config := fmt.Sprintf(`
			data "github_rest_api" "test" {
				endpoint = "/%v"
			}
		`, endpoint)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("Error: GET https://api.github.com/xx.*: 414"),
				},
			},
		})
	})
}
