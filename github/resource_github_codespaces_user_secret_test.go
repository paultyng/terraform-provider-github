package github

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubCodespacesUserSecret(t *testing.T) {
	t.Run("creates and updates secrets without error", func(t *testing.T) {
		secretValue := base64.StdEncoding.EncodeToString([]byte("super_secret_value"))
		updatedSecretValue := base64.StdEncoding.EncodeToString([]byte("updated_super_secret_value"))

		config := fmt.Sprintf(`
			resource "github_codespaces_user_secret" "plaintext_secret" {
			  secret_name      = "test_plaintext_secret"
			  plaintext_value  = "%s"
			}

			resource "github_codespaces_user_secret" "encrypted_secret" {
			  secret_name      = "test_encrypted_secret"
			  encrypted_value  = "%s"
			}
		`, secretValue, secretValue)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_codespaces_user_secret.plaintext_secret", "plaintext_value",
					secretValue,
				),
				resource.TestCheckResourceAttr(
					"github_codespaces_user_secret.encrypted_secret", "encrypted_value",
					secretValue,
				),
				resource.TestCheckResourceAttrSet(
					"github_codespaces_user_secret.plaintext_secret", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_codespaces_user_secret.plaintext_secret", "updated_at",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_codespaces_user_secret.plaintext_secret", "plaintext_value",
					updatedSecretValue,
				),
				resource.TestCheckResourceAttr(
					"github_codespaces_user_secret.encrypted_secret", "encrypted_value",
					updatedSecretValue,
				),
				resource.TestCheckResourceAttrSet(
					"github_codespaces_user_secret.plaintext_secret", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_codespaces_user_secret.plaintext_secret", "updated_at",
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, individual) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks["before"],
				},
				{
					Config: strings.Replace(config,
						secretValue,
						updatedSecretValue, 2),
					Check: checks["after"],
				},
			},
		})
	})

	t.Run("deletes secrets without error", func(t *testing.T) {
		config := `
				resource "github_codespaces_user_secret" "plaintext_secret" {
					secret_name      = "test_plaintext_secret"
				}

				resource "github_codespaces_user_secret" "encrypted_secret" {
					secret_name      = "test_encrypted_secret"
				}
			`

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, individual) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:  config,
					Destroy: true,
				},
			},
		})
	})

	t.Run("imports secrets without error", func(t *testing.T) {
		secretValue := "super_secret_value"

		config := fmt.Sprintf(`
			resource "github_codespaces_user_secret" "test_secret" {
				secret_name      = "test_plaintext_secret"
				plaintext_value  = "%s"
			}
		`, secretValue)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_codespaces_user_secret.test_secret", "plaintext_value",
				secretValue,
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, individual) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:            "github_codespaces_user_secret.test_secret",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"plaintext_value"},
				},
			},
		})
	})
}
