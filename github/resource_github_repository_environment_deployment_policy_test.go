package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubRepositoryEnvironmentDeploymentPolicy(t *testing.T) {
	t.Run("creates a repository environment with branch-based deployment policy", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment / test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	   = github_repository.test.name
				environment	   = github_repository_environment.test.environment
				branch_pattern = "releases/*"
			}

		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment / test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
				"releases/*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
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

	t.Run("updates the pattern for a branch-based deployment policy", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		var deploymentPolicyId string

		config1 := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				branch_pattern = "main"
			}

		`, randomID)

		check1 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
				"main",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
			),
			testDeploymentPolicyId("github_repository_environment_deployment_policy.test", &deploymentPolicyId),
		)

		config2 := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				branch_pattern = "release/*"
			}

		`, randomID)

		check2 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
				"release/*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
			),
			testSameDeploymentPolicyId(
				"github_repository_environment_deployment_policy.test",
				&deploymentPolicyId,
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config1,
					Check:  check1,
				},
				{
					Config: config2,
					Check:  check2,
				},
			},
		})
	})

	t.Run("creates a repository environment with tag-based deployment policy", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				tag_pattern = "v*"
			}

		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
				"v*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
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

	t.Run("updates the pattern for a tag-based deployment policy", func(t *testing.T) {
		var deploymentPolicyId string
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config1 := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				tag_pattern = "v*"
			}

		`, randomID)

		check1 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
				"v*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
			),
			testDeploymentPolicyId("github_repository_environment_deployment_policy.test", &deploymentPolicyId),
		)

		config2 := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				tag_pattern = "version*"
			}

		`, randomID)

		check2 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
				"version*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
			),
			testSameDeploymentPolicyId(
				"github_repository_environment_deployment_policy.test",
				&deploymentPolicyId,
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config1,
					Check:  check1,
				},
				{
					Config: config2,
					Check:  check2,
				},
			},
		})
	})

	t.Run("recreates deployment policy when pattern type changes from branch to tag", func(t *testing.T) {
		var deploymentPolicyId string
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config1 := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				branch_pattern = "release/*"
			}

		`, randomID)

		check1 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
				"release/*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
			),
			testDeploymentPolicyId("github_repository_environment_deployment_policy.test", &deploymentPolicyId),
		)

		config2 := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				tag_pattern = "v*"
			}

		`, randomID)

		check2 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
				"v*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
			),
			testNewDeploymentPolicyId(
				"github_repository_environment_deployment_policy.test",
				&deploymentPolicyId,
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config1,
					Check:  check1,
				},
				{
					Config: config2,
					Check:  check2,
				},
			},
		})
	})

	t.Run("recreates deployment policy when pattern type changes from tag to branch", func(t *testing.T) {
		var deploymentPolicyId string
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		config1 := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				tag_pattern = "v*"
			}

		`, randomID)

		check1 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
				"v*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
			),
			testDeploymentPolicyId("github_repository_environment_deployment_policy.test", &deploymentPolicyId),
		)

		config2 := fmt.Sprintf(`

			data "github_user" "current" {
				username = ""
			}

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				ignore_vulnerability_alerts_during_read = true
			}

			resource "github_repository_environment" "test" {
				repository 	= github_repository.test.name
				environment	= "environment/test"
				wait_timer	= 10000
				reviewers {
					users = [data.github_user.current.id]
				}
				deployment_branch_policy {
					protected_branches     = false
					custom_branch_policies = true
				}
			}

			resource "github_repository_environment_deployment_policy" "test" {
				repository 	= github_repository.test.name
				environment	= github_repository_environment.test.environment
				branch_pattern = "release/*"
			}

		`, randomID)

		check2 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "repository",
				fmt.Sprintf("tf-acc-test-%s", randomID),
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "environment",
				"environment/test",
			),
			resource.TestCheckResourceAttr(
				"github_repository_environment_deployment_policy.test", "branch_pattern",
				"release/*",
			),
			resource.TestCheckNoResourceAttr(
				"github_repository_environment_deployment_policy.test", "tag_pattern",
			),
			testNewDeploymentPolicyId(
				"github_repository_environment_deployment_policy.test",
				&deploymentPolicyId,
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config1,
					Check:  check1,
				},
				{
					Config: config2,
					Check:  check2,
				},
			},
		})
	})
}

func testDeploymentPolicyId(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID is not set")
		}

		*id = rs.Primary.ID

		return nil
	}
}

func testSameDeploymentPolicyId(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID is not set")
		}

		if rs.Primary.ID != *id {
			return fmt.Errorf("New resource does not match old resource id: %s, %s", rs.Primary.ID, *id)
		}

		return nil
	}
}

func testNewDeploymentPolicyId(resourceName string, id *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource ID is not set")
		}

		if rs.Primary.ID == *id {
			return fmt.Errorf("New resource matches old resource id: %s", rs.Primary.ID)
		}

		return nil
	}
}
