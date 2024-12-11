package github

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubReleaseResource(t *testing.T) {
	t.Run("create a release with defaults", func(t *testing.T) {
		randomRepoPart := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		randomVersion := fmt.Sprintf("v1.0.%d", acctest.RandIntRange(0, 9999))

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_release" "test" {
			  repository 	   = github_repository.test.name
			  tag_name 		   = "%s"
			}
		`, randomRepoPart, randomVersion)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_release.test", "tag_name", randomVersion,
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "target_commitish", "main",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "name", "",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "body", "",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "draft", "true",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "prerelease", "true",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "generate_release_notes", "false",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "discussion_category_name", "",
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
					ResourceName:      "github_release.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: importReleaseByResourcePaths(
						"github_repository.test", "github_release.test"),
				},
			},
		})
	})

	t.Run("create a release on branch", func(t *testing.T) {
		randomRepoPart := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		randomVersion := fmt.Sprintf("v1.0.%d", acctest.RandIntRange(0, 9999))
		testBranchName := "test"

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_branch" "test" {
				repository    = github_repository.test.name
				branch        = "%s"
				source_branch = github_repository.test.default_branch
			}

			resource "github_release" "test" {
				repository 	   	 = github_repository.test.name
				tag_name 		 = "%s"
				target_commitish = github_branch.test.branch
			  	draft			 = false
				prerelease		 = false
			}
		`, randomRepoPart, testBranchName, randomVersion)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_release.test", "tag_name", randomVersion,
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "target_commitish", testBranchName,
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "name", "",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "body", "",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "draft", "false",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "prerelease", "false",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "generate_release_notes", "false",
			),
			resource.TestCheckResourceAttr(
				"github_release.test", "discussion_category_name", "",
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
					ResourceName:      "github_release.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: importReleaseByResourcePaths(
						"github_repository.test", "github_release.test"),
				},
			},
		})
	})
}

func importReleaseByResourcePaths(repoLogicalName, releaseLogicalName string) resource.ImportStateIdFunc {
	// test importing using an ID of the form <repo-node-id>:<release-id>
	// by retrieving the GraphQL ID from the terraform.State
	return func(s *terraform.State) (string, error) {
		log.Printf("[DEBUG] Looking up tf state ")
		repo := s.RootModule().Resources[repoLogicalName]
		if repo == nil {
			return "", fmt.Errorf("Cannot find %s in terraform state", repoLogicalName)
		}
		repoID := repo.Primary.ID
		if repoID == "" {
			return "", fmt.Errorf("repository %s does not have an id in terraform state", repoLogicalName)
		}

		release := s.RootModule().Resources[releaseLogicalName]
		if release == nil {
			return "", fmt.Errorf("Cannot find %s in terraform state", releaseLogicalName)
		}
		releaseID := release.Primary.ID
		if releaseID == "" {
			return "", fmt.Errorf("release %s does not have an id in terraform state", releaseLogicalName)
		}

		return fmt.Sprintf("%s:%s", repoID, releaseID), nil
	}
}
