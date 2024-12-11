package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryDataSource(t *testing.T) {
	t.Run("queries a public repository without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_repository" "test" {
				full_name = "%s/%s"
			}
		`, testAccConf.testPublicRepositoryOwner, testAccConf.testPublicRepository)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "full_name",
				fmt.Sprintf("%s/%s", testAccConf.testPublicRepositoryOwner, testAccConf.testPublicRepository)),
		)
		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("queries repository belonging to the current user without error", func(t *testing.T) {
		if len(testAccConf.testUserRepository) == 0 {
			t.Skip("No test user repository provided")
		}

		config := fmt.Sprintf(`
			data "github_repository" "test" {
				full_name = "%s/%s"
			}
		`, testAccConf.username, testAccConf.testUserRepository)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "full_name",
				fmt.Sprintf("%s/%s", testAccConf.username, testAccConf.testUserRepository)),
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

	t.Run("queries an org repository without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_repository" "test" {
				full_name = "%s/%s"
			}
		`, testAccConf.owner, testAccConf.testOrgRepository)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "full_name",
				fmt.Sprintf("%s/%s", testAccConf.owner, testAccConf.testOrgRepository)),
		)
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("queries a repository with pages configured", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name         = "tf-acc-%s"
				auto_init    = true
				pages {
					source {
						branch = "main"
					}
				}
			}

			data "github_repository" "test" {
				name = github_repository.test.name
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "pages.0.source.0.branch",
				"main",
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

	t.Run("checks defaults on a new repository", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-%s", randomID)

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name         = "%s"
				auto_init    = true
			}

			data "github_repository" "test" {
				name = github_repository.test.name
			}
		`, repoName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_repository.test", "name", repoName),
			resource.TestCheckResourceAttrSet("data.github_repository.test", "has_projects"),
			resource.TestCheckResourceAttr("data.github_repository.test", "description", ""),
			resource.TestCheckResourceAttr("data.github_repository.test", "homepage_url", ""),
			resource.TestCheckResourceAttr("data.github_repository.test", "pages.#", "0"),
			resource.TestCheckResourceAttr("data.github_repository.test", "fork", "false"),
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

	t.Run("queries a public repository that is a template", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_repository" "test" {
				full_name = "%s/%s"
			}
		`, testAccConf.testPublicTemplateRepositoryOwner, testAccConf.testPublicTemplateRepository)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "is_template",
				"true",
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

	t.Run("queries an org repository that is a template", func(t *testing.T) {
		if len(testAccConf.testOrgTemplateRepository) == 0 {
			t.Skip("No org template repository provided")
		}

		config := fmt.Sprintf(`
			data "github_repository" "test" {
				full_name = "%s/%s"
			}
		`, testAccConf.owner, testAccConf.testOrgTemplateRepository)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "is_template",
				"true",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("queries a repository that was generated from a template", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-%s"
				template {
					owner      = "%s"
					repository = "%s"
				}
			}

			data "github_repository" "test" {
				name = github_repository.test.name
			}
		`, randomID, testAccConf.testPublicTemplateRepositoryOwner, testAccConf.testPublicTemplateRepository)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "template.0.owner",
				"template-repository",
			),
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "template.0.repository",
				"template-repository",
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

	t.Run("queries a repository that has no primary_language", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-%s"
			}

			data "github_repository" "test" {
				name = github_repository.test.name
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "primary_language",
				"",
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

	t.Run("queries a repository that has go as primary_language", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-%s"
				auto_init = true
			}
			resource "github_repository_file" "test" {
				repository     = github_repository.test.name
				file           = "test.go"
				content        = "package main"
			}

			data "github_repository" "test" {
				name = github_repository_file.test.repository
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "primary_language",
				"Go",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					// Not doing any checks since the language doesnt have time to be updated on the first apply
					Config: config,
				},
				{
					// Re-running the terraform will refresh the language since the go-file has been created
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("queries a repository that has a license", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-%s"
				auto_init = true
			}
			resource "github_repository_file" "test" {
				repository     = github_repository.test.name
				file           = "LICENSE"
				content             = <<EOT

Copyright (c) 2011-2023 GitHub Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE."
EOT
}

			data "github_repository" "test" {
				name = github_repository_file.test.repository
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_repository.test", "repository_license.0.license.0.spdx_id",
				"MIT",
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
}
