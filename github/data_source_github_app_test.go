package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubAppDataSource(t *testing.T) {
	t.Run("queries a global app", func(t *testing.T) {
		config := `
			data "github_app" "test" {
				slug = "github-actions"
			}
		`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(
				"data.github_app.test", "name",
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
}
