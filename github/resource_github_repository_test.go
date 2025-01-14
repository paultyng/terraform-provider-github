package github

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccGithubRepositories(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates and updates repositories without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {

				name                        = "tf-acc-test-create-%[1]s"
				description                 = "Terraform acceptance tests %[1]s"
				has_discussions             = true
				has_issues                  = true
				has_wiki                    = true
				has_downloads               = true
				allow_merge_commit          = true
				allow_squash_merge          = false
				allow_rebase_merge          = false
				allow_auto_merge            = true
				allow_forking               = true
				merge_commit_title          = "MERGE_MESSAGE"
				merge_commit_message        = "PR_TITLE"
				auto_init                   = false
				web_commit_signoff_required = true
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.test", "has_issues",
				"true",
			),
			resource.TestCheckResourceAttr(
				"github_repository.test", "has_discussions",
				"true",
			),
			resource.TestCheckResourceAttr(
				"github_repository.test", "allow_auto_merge",
				"true",
			),
			resource.TestCheckResourceAttr(
				"github_repository.test", "allow_forking",
				"true",
			),
			resource.TestCheckResourceAttr(
				"github_repository.test", "merge_commit_title",
				"MERGE_MESSAGE",
			),
			resource.TestCheckResourceAttr(
				"github_repository.test", "web_commit_signoff_required",
				"true",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("updates a repositories name without error", func(t *testing.T) {

		oldName := fmt.Sprintf(`tf-acc-test-rename-%[1]s`, randomID)
		newName := fmt.Sprintf(`%[1]s-renamed`, oldName)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name         = "%[1]s"
			  description  = "Terraform acceptance tests %[2]s"
			}
		`, oldName, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "name",
					oldName,
				),
				resource.ComposeTestCheckFunc(
					testCheckResourceAttrContains("github_repository.test", "full_name",
						oldName),
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "name",
					newName,
				),
				resource.ComposeTestCheckFunc(
					testCheckResourceAttrContains("github_repository.test", "full_name",
						newName),
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						// Rename the repo to something else
						Config: strings.Replace(
							config,
							oldName,
							newName, 1),
						Check: checks["after"],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("imports repositories without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name         = "tf-acc-test-import-%[1]s"
			  description  = "Terraform acceptance tests %[1]s"
				auto_init 	 = false
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_repository.test", "name"),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
					{
						ResourceName:      "github_repository.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("archives repositories without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name         = "tf-acc-test-archive-%[1]s"
			  description  = "Terraform acceptance tests %[1]s"
				archived     = false
			}
		`, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "archived",
					"false",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "archived",
					"true",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						Config: strings.Replace(config,
							`archived     = false`,
							`archived     = true`, 1),
						Check: checks["after"],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("manages the project feature for a repository", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name         = "tf-acc-test-project-%[1]s"
			  description  = "Terraform acceptance tests %[1]s"
				has_projects = false
			}
		`, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "has_projects",
					"false",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "has_projects",
					"true",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						Config: strings.Replace(config,
							`has_projects = false`,
							`has_projects = true`, 1),
						Check: checks["after"],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("manages the default branch feature for a repository", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name           = "tf-acc-test-branch-%[1]s"
			  description    = "Terraform acceptance tests %[1]s"
			  default_branch = "main"
			  auto_init      = true
			}

			resource "github_branch" "default" {
			  repository = github_repository.test.name
			  branch     = "default"
			}
		`, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "default_branch",
					"main",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "default_branch",
					"default",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					// Test changing default_branch
					{
						Config: strings.Replace(config,
							`default_branch = "main"`,
							`default_branch = "default"`, 1),
						Check: checks["after"],
					},
					// Test changing default_branch back to main again
					{
						Config: config,
						Check:  checks["before"],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("allows setting default_branch on an empty repository", func(t *testing.T) {

		// Although default_branch is deprecated, for backwards compatibility
		// we allow setting it to "main".

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name           = "tf-acc-test-empty-%[1]s"
			  description    = "Terraform acceptance tests %[1]s"
			  default_branch = "main"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.test", "default_branch",
				"main",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					// Test creation with default_branch set
					{
						Config: config,
						Check:  check,
					},
					// Test that changing another property does not try to set
					// default_branch (which would crash).
					{
						Config: strings.Replace(config,
							`acceptance tests`,
							`acceptance test`, 1),
						Check: check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("manages the license and gitignore feature for a repository", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name           = "tf-acc-test-license-%[1]s"
				description    = "Terraform acceptance tests %[1]s"
				license_template   = "ms-pl"
				gitignore_template = "C++"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.test", "license_template",
				"ms-pl",
			),
			resource.TestCheckResourceAttr(
				"github_repository.test", "gitignore_template",
				"C++",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("configures topics for a repository", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-topic-%[1]s"
				description = "Terraform acceptance tests %[1]s"
				topics			= ["terraform", "testing"]
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.test", "topics.#",
				"2",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("creates a repository using a template", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name        = "tf-acc-test-template-%s"
				description = "Terraform acceptance tests %[1]s"

				template {
					owner = "%s"
					repository = "%s"
				}

			}
		`, randomID, testOrganization, "terraform-template-module")

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.test", "is_template",
				"false",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("archives repositories on destroy", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name               = "tf-acc-test-destroy-%[1]s"
				auto_init          = true
				archive_on_destroy = true
				archived           = false
			}
		`, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "archived",
					"false",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "archived",
					"true",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						Config: strings.Replace(config,
							`archived           = false`,
							`archived           = true`, 1),
						Check: checks["after"],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("configures vulnerability alerts", func(t *testing.T) {

		t.Run("for a public repository", func(t *testing.T) {

			config := fmt.Sprintf(`
				resource "github_repository" "test" {
					name       = "tf-acc-test-pub-vuln-%s"
					visibility = "public"
				}
			`, randomID)

			checks := map[string]resource.TestCheckFunc{
				"before": resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"github_repository.test", "vulnerability_alerts",
						"false",
					),
				),
				"after": resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"github_repository.test", "vulnerability_alerts",
						"true",
					),
					resource.TestCheckResourceAttr(
						"github_repository.test", "visibility",
						"public",
					),
				),
			}

			testCase := func(t *testing.T, mode string) {
				resource.Test(t, resource.TestCase{
					PreCheck:  func() { skipUnlessMode(t, mode) },
					Providers: testAccProviders,
					Steps: []resource.TestStep{
						{
							Config: config,
							Check:  checks["before"],
						},
						{
							Config: strings.Replace(config,
								`}`,
								"vulnerability_alerts = true\n}", 1),
							Check: checks["after"],
						},
					},
				})
			}

			t.Run("with an anonymous account", func(t *testing.T) {
				t.Skip("anonymous account not supported for this operation")
			})

			t.Run("with an individual account", func(t *testing.T) {
				testCase(t, individual)
			})

			t.Run("with an organization account", func(t *testing.T) {
				testCase(t, organization)
			})
		})

		t.Run("for a private repository", func(t *testing.T) {

			config := fmt.Sprintf(`
				resource "github_repository" "test" {
					name       = "tf-acc-test-prv-vuln-%s"
					visibility = "private"
				}
			`, randomID)

			checks := map[string]resource.TestCheckFunc{
				"before": resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"github_repository.test", "vulnerability_alerts",
						"false",
					),
				),
				"after": resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"github_repository.test", "vulnerability_alerts",
						"true",
					),
					resource.TestCheckResourceAttr(
						"github_repository.test", "visibility",
						"private",
					),
				),
			}

			testCase := func(t *testing.T, mode string) {
				resource.Test(t, resource.TestCase{
					PreCheck:  func() { skipUnlessMode(t, mode) },
					Providers: testAccProviders,
					Steps: []resource.TestStep{
						{
							Config: config,
							Check:  checks["before"],
						},
						{
							Config: strings.Replace(config,
								`}`,
								"vulnerability_alerts = true\n}", 1),
							Check: checks["after"],
						},
					},
				})
			}

			t.Run("with an anonymous account", func(t *testing.T) {
				t.Skip("anonymous account not supported for this operation")
			})

			t.Run("with an individual account", func(t *testing.T) {
				testCase(t, individual)
			})

			t.Run("with an organization account", func(t *testing.T) {
				testCase(t, organization)
			})
		})

	})

	t.Run("create and modify merge commit strategy without error", func(t *testing.T) {
		mergeCommitTitle := "PR_TITLE"
		mergeCommitMessage := "BLANK"
		updatedMergeCommitTitle := "MERGE_MESSAGE"
		updatedMergeCommitMessage := "PR_TITLE"

		configs := map[string]string{
			"before": fmt.Sprintf(`
                		resource "github_repository" "test" {

		                	name                 = "tf-acc-test-modify-co-str-%[1]s"
		                  	allow_merge_commit   = true
		                  	merge_commit_title   = "%s"
		                  	merge_commit_message = "%s"
		                }
		        `, randomID, mergeCommitTitle, mergeCommitMessage),
			"after": fmt.Sprintf(`
		                resource "github_repository" "test" {
		                  	name                 = "tf-acc-test-modify-co-str-%[1]s"
		                  	allow_merge_commit   = true
		                  	merge_commit_title   = "%s"
		                  	merge_commit_message = "%s"
		                }
		        `, randomID, updatedMergeCommitTitle, updatedMergeCommitMessage),
		}

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "merge_commit_title",
					mergeCommitTitle,
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "merge_commit_message",
					mergeCommitMessage,
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "merge_commit_title",
					updatedMergeCommitTitle,
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "merge_commit_message",
					updatedMergeCommitMessage,
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
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
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("create and modify squash merge commit strategy without error", func(t *testing.T) {
		squashMergeCommitTitle := "PR_TITLE"
		squashMergeCommitMessage := "PR_BODY"
		updatedSquashMergeCommitTitle := "COMMIT_OR_PR_TITLE"
		updatedSquashMergeCommitMessage := "COMMIT_MESSAGES"

		configs := map[string]string{
			"before": fmt.Sprintf(`
	                	resource "github_repository" "test" {
	                  		name                        = "tf-acc-test-modify-sq-str-%[1]s"
	                  		allow_squash_merge          = true
	                  		squash_merge_commit_title   = "%s"
	                  		squash_merge_commit_message = "%s"
	                	}
	            	`, randomID, squashMergeCommitTitle, squashMergeCommitMessage),
			"after": fmt.Sprintf(`
	                	resource "github_repository" "test" {
	                  		name                        = "tf-acc-test-modify-sq-str-%[1]s"
	                  		allow_squash_merge          = true
	                  		squash_merge_commit_title   = "%s"
	                  		squash_merge_commit_message = "%s"
	                	}
	            	`, randomID, updatedSquashMergeCommitTitle, updatedSquashMergeCommitMessage),
		}

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "squash_merge_commit_title",
					squashMergeCommitTitle,
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "squash_merge_commit_message",
					squashMergeCommitMessage,
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "squash_merge_commit_title",
					updatedSquashMergeCommitTitle,
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "squash_merge_commit_message",
					updatedSquashMergeCommitMessage,
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
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
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("create a repository with go as primary_language", func(t *testing.T) {
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
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.test", "primary_language",
				"Go",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						// Not doing any checks since the file needs to be created before the language can be updated
						Config: config,
					},
					{
						// Re-running the terraform will refresh the language since the go-file has been created
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

}
func TestAccGithubRepositoryPages(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("manages the legacy pages feature for a repository", func(t *testing.T) {

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
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.test", "pages.0.source.0.branch",
				"main",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("manages the pages from workflow feature for a repository", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name         = "tf-acc-%s"
				auto_init    = true
				pages {
					build_type = "workflow"
				}
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.test", "pages.0.source.0.branch",
				"main",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("expand Pages configuration with workflow", func(t *testing.T) {
		input := []interface{}{map[string]interface{}{
			"build_type": "workflow",
			"source":     []interface{}{map[string]interface{}{}},
		}}

		pages := expandPages(input)
		if pages == nil {
			t.Fatal("pages is nil")
		}
		if pages.GetBuildType() != "workflow" {
			t.Errorf("got %q; want %q", pages.GetBuildType(), "workflow")
		}
		if pages.GetSource().GetBranch() != "main" {
			t.Errorf("got %q; want %q", pages.GetSource().GetBranch(), "main")
		}
	})

	t.Run("expand Pages configuration with source", func(t *testing.T) {
		input := []interface{}{map[string]interface{}{
			"build_type": "legacy",
			"source": []interface{}{map[string]interface{}{
				"branch": "main",
				"path":   "/docs",
			}},
		}}

		pages := expandPages(input)
		if pages == nil {
			t.Fatal("pages is nil")
		}
		if pages.GetBuildType() != "legacy" {
			t.Errorf("got %q; want %q", pages.GetBuildType(), "legacy")
		}
		if pages.GetSource().GetBranch() != "main" {
			t.Errorf("got %q; want %q", pages.GetSource().GetBranch(), "main")
		}
		if pages.GetSource().GetPath() != "/docs" {
			t.Errorf("got %q; want %q", pages.GetSource().GetPath(), "/docs")
		}
	})
}

func TestAccGithubRepositorySecurity(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("manages the security feature for a repository", func(t *testing.T) {

		t.Run("for a private repository", func(t *testing.T) {
			t.Skip("organization/individual must have purchased Advanced Security in order to enable it")

			config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name        = "tf-acc-%s"
			  description = "A repository created by Terraform to test security features"
			  visibility  = "private"
			  security_and_analysis {
			    advanced_security {
			      status = "enabled"
			    }
			    secret_scanning {
			      status = "enabled"
			    }
			    secret_scanning_push_protection {
			       status = "enabled"
			    }
			  }
			}
			`, randomID)

			check := resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "security_and_analysis.0.advanced_security.0.status",
					"enabled",
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "security_and_analysis.0.secret_scanning.0.status",
					"enabled",
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "security_and_analysis.0.secret_scanning_push_protection.0.status",
					"disabled",
				),
			)
			testCase := func(t *testing.T, mode string) {
				resource.Test(t, resource.TestCase{
					PreCheck:  func() { skipUnlessMode(t, mode) },
					Providers: testAccProviders,
					Steps: []resource.TestStep{
						{
							Config: config,
							Check:  check,
						},
					},
				})
			}
			t.Run("with an anonymous account", func(t *testing.T) {
				t.Skip("anonymous account not supported for this operation")
			})

			t.Run("with an individual account", func(t *testing.T) {
				testCase(t, individual)
			})

			t.Run("with an organization account", func(t *testing.T) {
				testCase(t, organization)
			})
		})

		t.Run("for a public repository", func(t *testing.T) {

			config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name        = "tf-acc-%s"
			  description = "A repository created by Terraform to test security features"
			  visibility  = "public"
			  security_and_analysis {
			    secret_scanning {
			      status = "enabled"
			    }
			    # seems like it can only be "enabled" for an organization that has purchased GHAS
			    secret_scanning_push_protection {
			       status = "disabled"
			    }
			  }
			}
			`, randomID)

			check := resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "security_and_analysis.0.secret_scanning.0.status",
					"enabled",
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "security_and_analysis.0.secret_scanning_push_protection.0.status",
					"disabled",
				),
			)
			testCase := func(t *testing.T, mode string) {
				resource.Test(t, resource.TestCase{
					PreCheck:  func() { skipUnlessMode(t, mode) },
					Providers: testAccProviders,
					Steps: []resource.TestStep{
						{
							Config: config,
							Check:  check,
						},
					},
				})
			}

			t.Run("with an anonymous account", func(t *testing.T) {
				t.Skip("anonymous account not supported for this operation")
			})

			t.Run("with an individual account", func(t *testing.T) {
				testCase(t, individual)
			})

			t.Run("with an organization account", func(t *testing.T) {
				testCase(t, organization)
			})
		})
	})
}

func TestAccGithubRepositoryVisibility(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates repos with private visibility", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "private" {
				name       = "tf-acc-test-visibility-private-%s"
				visibility = "private"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.private", "visibility",
				"private",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("creates repos with internal visibility", func(t *testing.T) {
		t.Skip("organization used in automated tests does not support internal repositories")

		config := fmt.Sprintf(`
			resource "github_repository" "internal" {
				name       = "tf-acc-test-visibility-internal-%s"
				visibility = "internal"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.internal", "visibility",
				"internal",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("updates repos to private visibility", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "public" {
				name       = "tf-acc-test-visibility-public-%s"
				visibility = "public"
				vulnerability_alerts = false
			}
		`, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.public", "visibility",
					"public",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.public", "visibility",
					"private",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						Config: reconfigureVisibility(config, "private"),
						Check:  checks["after"],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("updates repos to public visibility", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "tf-acc-test-prv-vuln-%s"
				visibility = "private"
			}
		`, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "vulnerability_alerts",
					"false",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "vulnerability_alerts",
					"true",
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "visibility",
					"private",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						Config: strings.Replace(config,
							`}`,
							"vulnerability_alerts = true\n}", 1),
						Check: checks["after"],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("updates repos to internal visibility", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "tf-acc-test-prv-vuln-%s"
				visibility = "private"
			}
		`, randomID)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "vulnerability_alerts",
					"false",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository.test", "vulnerability_alerts",
					"true",
				),
				resource.TestCheckResourceAttr(
					"github_repository.test", "visibility",
					"private",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checks["before"],
					},
					{
						Config: strings.Replace(config,
							`}`,
							"vulnerability_alerts = true\n}", 1),
						Check: checks["after"],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("sets private visibility for repositories created by a template", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "private" {
				name       = "tf-acc-test-visibility-private-%s"
				visibility = "private"
				template {
					owner      = "%s"
					repository = "%s"
				}
			}
		`, randomID, testOrganization, "terraform-template-module")

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository.private", "visibility",
				"private",
			),
			resource.TestCheckResourceAttr(
				"github_repository.private", "private",
				"true",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

}

func TestGithubRepositoryTopicPassesValidation(t *testing.T) {
	resource := resourceGithubRepository()
	schema := resource.Schema["topics"].Elem.(*schema.Schema)
	diags := schema.ValidateDiagFunc("ef69e1a3-66be-40ca-bb62-4f36186aa292", cty.Path{cty.GetAttrStep{Name: "topic"}})
	if diags.HasError() {
		t.Error(fmt.Errorf("unexpected topic validation failure: %s", diags[0].Summary))
	}
}

func TestGithubRepositoryTopicFailsValidationWhenOverMaxCharacters(t *testing.T) {
	resource := resourceGithubRepository()
	schema := resource.Schema["topics"].Elem.(*schema.Schema)

	diags := schema.ValidateDiagFunc(strings.Repeat("a", 51), cty.Path{cty.GetAttrStep{Name: "topic"}})
	if len(diags) != 1 {
		t.Error(fmt.Errorf("unexpected number of topic validation failures; expected=1; actual=%d", len(diags)))
	}
	expectedFailure := "invalid value for topics (must include only lowercase alphanumeric characters or hyphens and cannot start with a hyphen and consist of 50 characters or less)"
	actualFailure := diags[0].Summary
	if expectedFailure != actualFailure {
		t.Error(fmt.Errorf("unexpected topic validation failure; expected=%s; action=%s", expectedFailure, actualFailure))
	}
}

func testSweepRepositories(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	repos, _, err := client.Repositories.ListByUser(context.TODO(), meta.(*Owner).name, nil)
	if err != nil {
		return err
	}

	for _, r := range repos {
		if name := r.GetName(); strings.HasPrefix(name, "tf-acc-") || strings.HasPrefix(name, "foo-") {
			log.Printf("[DEBUG] Destroying Repository %s", name)

			if _, err := client.Repositories.Delete(context.TODO(), meta.(*Owner).name, name); err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	resource.AddTestSweepers("github_repository", &resource.Sweeper{
		Name: "github_repository",
		F:    testSweepRepositories,
	})
}

func reconfigureVisibility(config, visibility string) string {
	re := regexp.MustCompile(`visibility = "(.*)"`)
	newConfig := re.ReplaceAllString(
		config,
		fmt.Sprintf(`visibility = "%s"`, visibility),
	)
	return newConfig
}

type resourceDataLike map[string]interface{}

func (d resourceDataLike) GetOk(key string) (interface{}, bool) {
	v, ok := d[key]
	return v, ok
}

func TestResourceGithubParseFullName(t *testing.T) {
	repo, org, ok := resourceGithubParseFullName(resourceDataLike(map[string]interface{}{"full_name": "myrepo/myorg"}))
	assert.True(t, ok)
	assert.Equal(t, "myrepo", repo)
	assert.Equal(t, "myorg", org)
	_, _, ok = resourceGithubParseFullName(resourceDataLike(map[string]interface{}{}))
	assert.False(t, ok)
	_, _, ok = resourceGithubParseFullName(resourceDataLike(map[string]interface{}{"full_name": "malformed"}))
	assert.False(t, ok)
}

func testCheckResourceAttrContains(resourceName, attributeName, substring string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		value, ok := rs.Primary.Attributes[attributeName]
		if !ok {
			return fmt.Errorf("Attribute not found: %s", attributeName)
		}

		if !strings.Contains(value, substring) {
			return fmt.Errorf("Attribute '%s' does not contain '%s'", value, substring)
		}

		return nil
	}
}

func TestGithubRepositoryNameFailsValidationWhenOverMaxCharacters(t *testing.T) {
	resource := resourceGithubRepository()
	schema := resource.Schema["name"]

	diags := schema.ValidateDiagFunc(strings.Repeat("a", 101), cty.GetAttrPath("name"))
	if len(diags) != 1 {
		t.Error(fmt.Errorf("unexpected number of name validation failures; expected=1; actual=%d", len(diags)))
	}
	expectedFailure := "invalid value for name (must include only alphanumeric characters, underscores or hyphens and consist of 100 characters or less)"
	actualFailure := diags[0].Summary
	if expectedFailure != actualFailure {
		t.Error(fmt.Errorf("unexpected name validation failure; expected=%s; action=%s", expectedFailure, actualFailure))
	}
}

func TestGithubRepositoryNameFailsValidationWithSpace(t *testing.T) {
	resource := resourceGithubRepository()
	schema := resource.Schema["name"]

	diags := schema.ValidateDiagFunc("test space", cty.GetAttrPath("name"))
	if len(diags) != 1 {
		t.Error(fmt.Errorf("unexpected number of name validation failures; expected=1; actual=%d", len(diags)))
	}
	expectedFailure := "invalid value for name (must include only alphanumeric characters, underscores or hyphens and consist of 100 characters or less)"
	actualFailure := diags[0].Summary
	if expectedFailure != actualFailure {
		t.Error(fmt.Errorf("unexpected name validation failure; expected=%s; action=%s", expectedFailure, actualFailure))
	}
}
