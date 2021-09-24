package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccGithubUsersDataSource(t *testing.T) {

	t.Run("queries multiple accounts", func(t *testing.T) {

		config := fmt.Sprintf(`
			data "github_users" "test" {
				usernames = ["%[1]s", "!%[1]s"]
			}
		`, testOwnerFunc())

		check := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_users.test", "logins.#", "1"),
			resource.TestCheckResourceAttr("data.github_users.test", "logins.0", testOwnerFunc()),
			resource.TestCheckResourceAttr("data.github_users.test", "node_ids.#", "1"),
			resource.TestCheckResourceAttr("data.github_users.test", "unknown_logins.#", "1"),
			resource.TestCheckResourceAttr("data.github_users.test", "unknown_logins.0", fmt.Sprintf("!%s", testOwnerFunc())),
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
