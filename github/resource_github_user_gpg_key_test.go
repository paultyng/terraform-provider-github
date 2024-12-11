package github

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubUserGpgKey(t *testing.T) {
	t.Run("creates a GPG key without error", func(t *testing.T) {
		keyPath := strings.ReplaceAll(filepath.Join("test-fixtures", "gpg-pubkey.asc"), "\\", "/")

		config := fmt.Sprintf(`
		resource "github_user_gpg_key" "test" {
			armored_public_key = "${file("%s")}"
		}
		`, keyPath)

		check := resource.ComposeTestCheckFunc(
			resource.TestMatchResourceAttr(
				"github_user_gpg_key.test",
				"armored_public_key",
				regexp.MustCompile("^-----BEGIN PGP PUBLIC KEY BLOCK-----"),
			),
			resource.TestCheckResourceAttr(
				"github_user_gpg_key.test",
				"key_id",
				"AC541D2D1709CD33",
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
