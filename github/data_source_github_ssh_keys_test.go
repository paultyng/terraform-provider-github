package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubSshKeysDataSource(t *testing.T) {
	t.Run("reads SSH keys without error", func(t *testing.T) {
		config := `data "github_ssh_keys" "test" {}`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_ssh_keys.test", "keys.#"),
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
