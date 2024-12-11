package github

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubBranchProtectionV4(t *testing.T) {
	t.Run("configures default settings when empty", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

		resource "github_repository" "test" {
		  name      = "tf-acc-test-%s"
		  auto_init = true
		}

		resource "github_branch_protection" "test" {

		  repository_id = github_repository.test.node_id
		  pattern       = "main"

		}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "pattern", "main",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "require_signed_commits", "false",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "require_conversation_resolution", "false",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_status_checks.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "lock_branch", "false",
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
				{
					ResourceName:      "github_branch_protection.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: importBranchProtectionByRepoName(
						fmt.Sprintf("tf-acc-test-%s", randomID), "main",
					),
				},
				{
					ResourceName: "github_branch_protection.test",
					ImportState:  true,
					ExpectError: regexp.MustCompile(
						`could not find a branch protection rule with the pattern 'no-such-pattern'`,
					),
					ImportStateIdFunc: importBranchProtectionByRepoName(
						fmt.Sprintf("tf-acc-test-%s", randomID), "no-such-pattern",
					),
				},
			},
		})
	})

	t.Run("configures default settings when conversation resolution is true", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

		resource "github_repository" "test" {
		  name      = "tf-acc-test-%s"
		  auto_init = true
		}

		resource "github_branch_protection" "test" {

		  repository_id = github_repository.test.node_id
		  pattern       = "main"

		  require_conversation_resolution = true
		}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "pattern", "main",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "require_signed_commits", "false",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "require_conversation_resolution", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_linear_history", "false",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_status_checks.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.#", "0",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "lock_branch", "false",
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
				{
					ResourceName:      "github_branch_protection.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: importBranchProtectionByRepoName(
						fmt.Sprintf("tf-acc-test-%s", randomID), "main",
					),
				},
				{
					ResourceName: "github_branch_protection.test",
					ImportState:  true,
					ExpectError: regexp.MustCompile(
						`could not find a branch protection rule with the pattern 'no-such-pattern'`,
					),
					ImportStateIdFunc: importBranchProtectionByRepoName(
						fmt.Sprintf("tf-acc-test-%s", randomID), "no-such-pattern",
					),
				},
			},
		})
	})

	t.Run("configures required status checks", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  required_status_checks {
			    strict   = true
			    contexts = ["github/foo"]
			  }

			}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_status_checks.#", "1",
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
				{
					ResourceName:      "github_branch_protection.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: importBranchProtectionByRepoID(
						"github_repository.test", "main"),
				},
				{
					ResourceName: "github_branch_protection.test",
					ImportState:  true,
					ExpectError: regexp.MustCompile(
						`could not find a branch protection rule with the pattern 'no-such-pattern'`,
					),
					ImportStateIdFunc: importBranchProtectionByRepoID(
						"github_repository.test", "no-such-pattern"),
				},
			},
		})
	})

	t.Run("configures required pull request reviews", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  required_pull_request_reviews {
				dismiss_stale_reviews      = true
				require_code_owner_reviews = true
				require_last_push_approval = true
			  }

			}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.0.dismiss_stale_reviews", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.0.require_code_owner_reviews", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.0.required_approving_review_count", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.0.require_last_push_approval", "true",
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

	t.Run("configures branch push restrictions", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  restrict_pushes {}
			}
	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.0.blocks_creations", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.0.push_allowances.#", "0",
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

	t.Run("configures branch push restrictions with node_id", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			data "github_user" "test" {
			  username = "%s"
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  restrict_pushes {
				push_allowances = [
					data.github_user.test.node_id,
				]
			  }
			}
	`, randomID, testAccConf.username)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.0.blocks_creations", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.0.push_allowances.#", "1",
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

	t.Run("configures branch push restrictions with username", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  restrict_pushes {
				push_allowances = [
					"/%s",
				]
			  }
			}
	`, randomID, testAccConf.username)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.0.blocks_creations", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.0.push_allowances.#", "1",
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

	t.Run("configures branch push restrictions with blocksCreations false", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  restrict_pushes {
				blocks_creations = false
			  }
			}
	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.0.blocks_creations", "false",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "restrict_pushes.0.push_allowances.#", "0",
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

	t.Run("configures force pushes and deletions", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			data "github_user" "test" {
			  username = "%s"
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  allows_deletions    = true
			  allows_force_pushes = true

			}
	`, randomID, testAccConf.username)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "allows_deletions", "true",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "allows_force_pushes", "true",
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

	t.Run("configures non-empty list of force push bypassers", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			data "github_user" "test" {
			  username = "%s"
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  force_push_bypassers = [
				data.github_user.test.node_id
			  ]

			}

	`, randomID, testAccConf.username)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "force_push_bypassers.#", "1",
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

	t.Run("configures allow force push with a team as bypasser", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_team" "test" {
				name = "tf-acc-test-%s"
			}

			resource "github_team_repository" "test" {
				team_id    = github_team.test.id
				repository = github_repository.test.name
				permission = "admin"
			}

			resource "github_branch_protection" "test" {
			  repository_id  = github_repository.test.node_id
			  pattern        = "main"

			  force_push_bypassers = [
				  "%s/${github_team.test.slug}"
	]
			}

	`, randomID, randomID, testAccConf.owner)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "force_push_bypassers.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "allows_force_pushes", "false",
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

	t.Run("configures empty list of force push bypassers", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  force_push_bypassers = []

			}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "force_push_bypassers.#", "0",
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

	t.Run("configures non-empty list of pull request bypassers", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection" "test" {

			  repository_id = github_repository.test.node_id
			  pattern       = "main"

			  required_pull_request_reviews {
				pull_request_bypassers = [
					"/%s",
				]
			  }

			}

			`, randomID, testAccConf.username)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.0.pull_request_bypassers.#", "1",
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

	t.Run("configures empty list of pull request bypassers", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_branch_protection" "test" {

			  repository_id  = github_repository.test.node_id
			  pattern        = "main"

			  required_pull_request_reviews {
				pull_request_bypassers = []
			  }

			}

	`, randomID)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_branch_protection.test", "required_pull_request_reviews.0.pull_request_bypassers.#", "0",
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

func importBranchProtectionByRepoName(repo, pattern string) resource.ImportStateIdFunc {
	// test importing using an ID of the form <repo-name>:<branch-protection-pattern>
	return func(s *terraform.State) (string, error) {
		return fmt.Sprintf("%s:%s", repo, pattern), nil
	}
}

func importBranchProtectionByRepoID(repoLogicalName, pattern string) resource.ImportStateIdFunc {
	// test importing using an ID of the form <repo-node-id>:<branch-protection-pattern>
	// by retrieving the GraphQL ID from the terraform.State
	return func(s *terraform.State) (string, error) {
		repo := s.RootModule().Resources[repoLogicalName]
		if repo == nil {
			return "", fmt.Errorf("cannot find %s in terraform state", repoLogicalName)
		}
		repoID, found := repo.Primary.Attributes["node_id"]
		if !found {
			return "", fmt.Errorf("repository %s does not have a node_id in terraform state", repo.Primary.ID)
		}
		return fmt.Sprintf("%s:%s", repoID, pattern), nil
	}
}

func testGithubBranchProtectionStateDataV1() map[string]interface{} {
	return map[string]interface{}{
		"blocks_creations":  true,
		"push_restrictions": [...]string{"/example-user"},
	}
}

func testGithubBranchProtectionStateDataV2() map[string]interface{} {
	restrictions := []interface{}{map[string]interface{}{
		"blocks_creations": true,
		"push_allowances":  [...]string{"/example-user"},
	}}
	return map[string]interface{}{
		"restrict_pushes": restrictions,
	}
}

func TestAccGithubBranchProtectionV4StateUpgradeV1(t *testing.T) {
	expected := testGithubBranchProtectionStateDataV2()
	actual, err := resourceGithubBranchProtectionUpgradeV1(context.Background(), testGithubBranchProtectionStateDataV1(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
