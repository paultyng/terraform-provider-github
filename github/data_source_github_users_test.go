package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TODO: this is failing
func TestAccGithubUsersDataSource(t *testing.T) {
	if len(testAccConf.testExternalUser) == 0 {
		t.Skip("No external user provided")
	}

	t.Run("queries multiple accounts", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_users" "test" {
				usernames = ["%[1]s", "!%[1]s"]
			}
		`, testAccConf.testExternalUser)

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_users.test", "logins.#", "1"),
			resource.TestCheckResourceAttr("data.github_users.test", "logins.0", testAccConf.testExternalUser),
			resource.TestCheckResourceAttr("data.github_users.test", "node_ids.#", "1"),
			resource.TestCheckResourceAttr("data.github_users.test", "unknown_logins.#", "1"),
			resource.TestCheckResourceAttr("data.github_users.test", "unknown_logins.0", fmt.Sprintf("!%s", testAccConf.testExternalUser)),
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

	t.Run("does not fail if called with empty list of usernames", func(t *testing.T) {
		config := `
			data "github_users" "test" {
				usernames = []
			}
		`

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_users.test", "logins.#", "0"),
			resource.TestCheckResourceAttr("data.github_users.test", "node_ids.#", "0"),
			resource.TestCheckResourceAttr("data.github_users.test", "unknown_logins.#", "0"),
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
