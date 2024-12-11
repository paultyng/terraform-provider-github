package github

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubBranchDataSource(t *testing.T) {
	t.Run("queries an existing branch without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%[1]s"
				auto_init = true
			}

			data "github_branch" "test" {
				repository = github_repository.test.name
				branch = "main"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestMatchResourceAttr(
				"data.github_branch.test", "ref", regexp.MustCompile("^refs/heads/main$"),
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

	// Can't test due to SDK and test framework limitations
	// t.Run("queries an invalid branch without error", func(t *testing.T) {
	// 	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	// 	config := fmt.Sprintf(`
	// 		resource "github_repository" "test" {
	// 		  name = "tf-acc-test-%[1]s"
	// 			auto_init = true
	// 		}

	// 		data "github_branch" "test" {
	// 			repository = github_repository.test.name
	// 			branch = "xxxxxx"
	// 		}
	// 	`, randomID)

	// 	check := resource.ComposeTestCheckFunc(
	// 		resource.TestCheckResourceAttr(
	// 			"data.github_branch.test", "ref", "",
	// 		),
	// 	)

	// 	resource.Test(t, resource.TestCase{
	// 		PreCheck:          func() { skipUnauthenticated(t) },
	// 		ProviderFactories: providerFactories,
	// 		Steps: []resource.TestStep{
	// 			{
	// 				Config: config,
	// 				Check:  check,
	// 			},
	// 		},
	// 	})
	// })
}
