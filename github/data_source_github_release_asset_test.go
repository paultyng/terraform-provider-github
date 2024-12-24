package github

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubReleaseAssetDataSource(t *testing.T) {

	testReleaseRepository := "go-github-issue-demo-1"
	if os.Getenv("GITHUB_TEMPLATE_REPOSITORY") != "" {
		testReleaseRepository = os.Getenv("GITHUB_TEMPLATE_REPOSITORY")
	}

	testReleaseAssetID := "151970555"
	if os.Getenv("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ASSET_ID") != "" {
		testReleaseAssetID = os.Getenv("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ASSET_ID")
	}

	testReleaseAssetName := "foo.txt"
	if os.Getenv("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ASSET_NAME") != "" {
		testReleaseAssetName = os.Getenv("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ASSET_NAME")
	}

	testReleaseAssetContent := "Hello, world!\n"
	if os.Getenv("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ASSET_CONTENT") != "" {
		testReleaseAssetContent = os.Getenv("GITHUB_TEMPLATE_REPOSITORY_RELEASE_ASSET_CONTENT")
	}

	testReleaseOwner := testOrganizationFunc()

	t.Run("queries specified asset ID", func(t *testing.T) {

		config := fmt.Sprintf(`
			data "github_release_asset" "test" {
				repository = "%s"
				owner = "%s"
				asset_id = "%s"
			}
		`, testReleaseRepository, testReleaseOwner, testReleaseAssetID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"data.github_release_asset.test", "asset_id", testReleaseAssetID,
			),
			resource.TestCheckResourceAttr(
				"data.github_release_asset.test", "name", testReleaseAssetName,
			),
			resource.TestCheckResourceAttr(
				"data.github_release_asset.test", "body", testReleaseAssetContent,
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
			testCase(t, anonymous)
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

}
