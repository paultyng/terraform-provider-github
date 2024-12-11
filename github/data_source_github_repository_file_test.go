package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryFileDataSource(t *testing.T) {
	t.Run("reads a file with a branch name provided without erroring", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true

				lifecycle {
					ignore_changes = all
				}
			}

			resource "github_repository_file" "test" {
				repository     = github_repository.test.name
				branch         = "main"
				file           = "test"
				content        = "bar"
				commit_message = "Managed by Terraform"
				commit_author  = "Terraform User"
				commit_email   = "terraform@example.com"
			}

			data "github_repository_file" "test" {
				repository     = github_repository.test.name
				branch         = "main"
				file           = github_repository_file.test.file
			}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "id"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "ref", "main"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "content", "bar"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "sha", "ba0e162e1c47469e3fe4b393a8bf8c569f302116"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_sha"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_message"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_author"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_email"),
					),
				},
			},
		})
	})

	t.Run("reads a file from a short repo name without erroring", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "%s"
				auto_init = true

				lifecycle {
					ignore_changes = all
				}
			}

			resource "github_repository_file" "test" {
				repository     = github_repository.test.name
				branch         = "main"
				file           = "test"
				content        = "bar"
				commit_message = "Managed by Terraform"
				commit_author  = "Terraform User"
				commit_email   = "terraform@example.com"
			}

			data "github_repository_file" "test" {
				repository     = "%[1]s"
				branch         = "main"
				file           = github_repository_file.test.file

				depends_on = [github_repository_file.test]
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "id"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "ref", "main"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "content", "bar"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "sha", "ba0e162e1c47469e3fe4b393a8bf8c569f302116"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_sha"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_message"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_author"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_email"),
					),
				},
			},
		})
	})

	t.Run("create and read a file without providing a branch name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name      			= "tf-acc-test-%s"
				auto_init 			= true

				lifecycle {
					ignore_changes = all
				}
			}

			resource "github_branch" "test" {
				repository = github_repository.test.name
				branch     = "test"
			}

			resource "github_branch_default" "default"{
				repository = github_repository.test.name
				branch     = github_branch.test.branch
			}

			resource "github_repository_file" "test" {
				repository     = github_repository.test.name
				branch         = github_branch_default.default.branch
				file           = "test"
				content        = "bar"
				commit_message = "Managed by Terraform"
				commit_author  = "Terraform User"
				commit_email   = "terraform@example.com"
			}

			data "github_repository_file" "test" {
				repository     = github_repository.test.name
				file           = github_repository_file.test.file
			}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "id"),
						resource.TestCheckNoResourceAttr("data.github_repository_file.test", "branch"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "ref", "test"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "content", "bar"),
						resource.TestCheckResourceAttr("data.github_repository_file.test", "sha", "ba0e162e1c47469e3fe4b393a8bf8c569f302116"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_sha"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_message"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_author"),
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "commit_email"),
					),
				},
			},
		})
	})

	// Can't test due to SDK and test framework limitations
	// t.Run("try reading a non-existent file without an error", func(t *testing.T) {
	// 	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	// 	config := fmt.Sprintf(`
	// 		resource "github_repository" "test" {
	// 			name      			= "tf-acc-test-%s"
	// 			auto_init 			= true

	// 			lifecycle {
	// 				ignore_changes = all
	// 			}
	// 		}

	// 		data "github_repository_file" "test" {
	// 			repository     = github_repository.test.name
	// 			file           = "test"
	// 		}
	// 	`, randomID)

	// 	resource.Test(t, resource.TestCase{
	// 		PreCheck:          func() { skipUnauthenticated(t) },
	// 		ProviderFactories: providerFactories,
	// 		Steps: []resource.TestStep{
	// 			{
	// 				Config: config,
	// 				Check: resource.ComposeTestCheckFunc(
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "id"),
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "branch"),
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "ref"),
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "content"),
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "sha"),
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "commit_sha"),
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "commit_message"),
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "commit_author"),
	// 					resource.TestCheckNoResourceAttr("data.github_repository_file.test", "commit_email"),
	// 				),
	// 			},
	// 		},
	// 	})
	// })

	t.Run("reads a directory without erroring", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      			= "tf-acc-test-%s"
				auto_init 			= true

				lifecycle {
					ignore_changes = all
				}
			}

			data "github_repository_file" "test" {
				repository     = github_repository.test.name
				file           = "."
			}
		`, randomID)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_repository_file.test", "id"),
						resource.TestCheckNoResourceAttr("data.github_repository_file.test", "ref"),
						resource.TestCheckNoResourceAttr("data.github_repository_file.test", "content"),
						resource.TestCheckNoResourceAttr("data.github_repository_file.test", "sha"),
						resource.TestCheckNoResourceAttr("data.github_repository_file.test", "commit_sha"),
						resource.TestCheckNoResourceAttr("data.github_repository_file.test", "commit_message"),
						resource.TestCheckNoResourceAttr("data.github_repository_file.test", "commit_author"),
						resource.TestCheckNoResourceAttr("data.github_repository_file.test", "commit_email"),
					),
				},
			},
		})
	})
}
