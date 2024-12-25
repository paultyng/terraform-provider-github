package github

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsOrganizationSecret(t *testing.T) {
	t.Run("creates and updates secrets without error", func(t *testing.T) {
		secretValue := base64.StdEncoding.EncodeToString([]byte("super_secret_value"))
		updatedSecretValue := base64.StdEncoding.EncodeToString([]byte("updated_super_secret_value"))

		config := fmt.Sprintf(`
			resource "github_actions_organization_secret" "plaintext_secret" {
			  secret_name      = "test_plaintext_secret"
			  plaintext_value  = "%s"
			  visibility       = "private"
			}

			resource "github_actions_organization_secret" "encrypted_secret" {
			  secret_name      = "test_encrypted_secret"
			  encrypted_value  = "%s"
			  visibility       = "private"
			  destroy_on_drift = false
			}
		`, secretValue, secretValue)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_organization_secret.plaintext_secret", "plaintext_value",
					secretValue,
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_secret.encrypted_secret", "encrypted_value",
					secretValue,
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_secret.plaintext_secret", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_secret.plaintext_secret", "updated_at",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_organization_secret.plaintext_secret", "plaintext_value",
					updatedSecretValue,
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_secret.encrypted_secret", "encrypted_value",
					updatedSecretValue,
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_secret.plaintext_secret", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_organization_secret.plaintext_secret", "updated_at",
				),
			),
		}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
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
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("deletes secrets without error", func(t *testing.T) {
		config := `
				resource "github_actions_organization_secret" "plaintext_secret" {
					secret_name      = "test_plaintext_secret"
					visibility       = "private"
				}

				resource "github_actions_organization_secret" "encrypted_secret" {
					secret_name      = "test_encrypted_secret"
					visibility       = "private"
				}
			`

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:  config,
						Destroy: true,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})

	t.Run("imports secrets without error", func(t *testing.T) {
		secretValue := "super_secret_value"

		config := fmt.Sprintf(`
			resource "github_actions_organization_secret" "test_secret" {
				secret_name      = "test_plaintext_secret"
				plaintext_value  = "%s"
				visibility       = "private"
			}
		`, secretValue)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_secret.test_secret", "plaintext_value",
				secretValue,
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
					{
						ResourceName:            "github_actions_organization_secret.test_secret",
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"plaintext_value"},
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			t.Skip("individual account not supported for this operation")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})
}
