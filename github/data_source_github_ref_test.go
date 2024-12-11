package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRefDataSource(t *testing.T) {
	t.Run("queries an existing branch ref without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%[1]s"
				auto_init = true
			}

			data "github_ref" "test" {
				owner      = "%s"
				repository = github_repository.test.name
				ref = "heads/main"
			}
		`, randomID, testAccConf.owner)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_ref.test", "id"),
						resource.TestCheckResourceAttrSet("data.github_ref.test", "sha"),
					),
				},
			},
		})
	})

	// Can't test due to SDK and test framework limitations
	// t.Run("queries an invalid ref without error", func(t *testing.T) {
	// 	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	// 	config := fmt.Sprintf(`
	// 		resource "github_repository" "test" {
	// 		  name = "tf-acc-test-%[1]s"
	// 			auto_init = true
	// 		}

	// 		data "github_ref" "test" {
	// 			repository = github_repository.test.id
	// 			ref = "heads/xxxxxx"
	// 		}
	// 	`, randomID)

	// 	check := resource.ComposeTestCheckFunc(
	// 		resource.TestCheckNoResourceAttr(
	// 			"data.github_ref.test", "id",
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
