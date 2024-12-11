package github

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubUserDataSource(t *testing.T) {
	if len(testAccConf.testExternalUser) == 0 {
		t.Skip("No external user provided")
	}

	t.Run("queries an existing individual account without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_user" "test" {
				username = "%s"
			}
		`, testAccConf.testExternalUser)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_user.test", "login"),
			resource.TestCheckResourceAttrSet("data.github_user.test", "id"),
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

	t.Run("errors when querying a non-existing individual account", func(t *testing.T) {
		config := fmt.Sprintf(`
				data "github_user" "test" {
					username = "!%s"
				}
			`, testAccConf.testExternalUser)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
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
