package github

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubReleaseDataSource(t *testing.T) {
	t.Run("queries latest release", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_release" "test" {
				repository = "%s"
				owner = "%s"
				retrieve_by = "latest"
			}
		`, testAccConf.testPublicRepository, testAccConf.testPublicRepositoryOwner)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_release.test", "id", strconv.Itoa(testAccConf.testPublicReleaseId),
			),
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

	t.Run("queries release by ID or tag", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_release" "by_id" {
				repository = "%[1]s"
				owner = "%[2]s"
				retrieve_by = "id"
				release_id = "%[3]d"
			}

			data "github_release" "by_tag" {
				repository = "%[1]s"
				owner = "%[2]s"
				retrieve_by = "tag"
				release_tag = data.github_release.by_id.release_tag
			}
		`, testAccConf.testPublicRepository, testAccConf.testPublicRepositoryOwner, testAccConf.testPublicReleaseId)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_release.by_id", "id", strconv.Itoa(testAccConf.testPublicReleaseId),
			),
			resource.TestCheckResourceAttr(
				"data.github_release.by_tag", "id", strconv.Itoa(testAccConf.testPublicReleaseId),
			),
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

	t.Run("errors when querying with non-existent ID", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_release" "test" {
				repository = "%s"
				owner = "%s"
				retrieve_by = "id"
			}
		`, testAccConf.testPublicRepository, testAccConf.testPublicRepositoryOwner)

		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("`release_id` must be set when `retrieve_by` = `id`"),
				},
			},
		})
	})

	t.Run("errors when querying with non-existent tag", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_release" "test" {
				repository = "%s"
				owner = "%s"
				retrieve_by = "tag"
			}
		`, testAccConf.testPublicRepository, testAccConf.testPublicRepositoryOwner)

		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile("`release_tag` must be set when `retrieve_by` = `tag`"),
				},
			},
		})
	})

	t.Run("errors when querying with non-existent repository", func(t *testing.T) {
		config := `
			data "github_release" "test" {
				repository = "test"
				owner = "test"
				retrieve_by = "latest"
			}
		`
		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Not Found`),
				},
			},
		})
	})
}
