package github

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubCodespacesOrganizationSecret(t *testing.T) {
	t.Run("creates and updates secrets without error", func(t *testing.T) {
		secretValue := base64.StdEncoding.EncodeToString([]byte("super_secret_value"))
		secretValueUpdated := base64.StdEncoding.EncodeToString([]byte("updated_super_secret_value"))

		config := fmt.Sprintf(`
			resource "github_codespaces_organization_secret" "plaintext_secret" {
			  secret_name      = "test_plaintext_secret"
			  plaintext_value  = "%[1]s"
			  visibility       = "private"
			}

			resource "github_codespaces_organization_secret" "encrypted_secret" {
			  secret_name      = "test_encrypted_secret"
			  encrypted_value  = "%[1]s"
			  visibility       = "private"
			}
		`, secretValue)

		configUpdated := fmt.Sprintf(`
			resource "github_codespaces_organization_secret" "plaintext_secret" {
			  secret_name      = "test_plaintext_secret"
			  plaintext_value  = "%[1]s"
			  visibility       = "private"
			}

			resource "github_codespaces_organization_secret" "encrypted_secret" {
			  secret_name      = "test_encrypted_secret"
			  encrypted_value  = "%[1]s"
			  visibility       = "private"
			}
		`, secretValueUpdated)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("github_codespaces_organization_secret.plaintext_secret", "plaintext_value", secretValue),
						resource.TestCheckResourceAttr("github_codespaces_organization_secret.encrypted_secret", "encrypted_value", secretValue),
						resource.TestCheckResourceAttrSet("github_codespaces_organization_secret.plaintext_secret", "created_at"),
						resource.TestCheckResourceAttrSet("github_codespaces_organization_secret.plaintext_secret", "updated_at"),
					),
				},
				{
					Config: configUpdated,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("github_codespaces_organization_secret.plaintext_secret", "plaintext_value", secretValueUpdated),
						resource.TestCheckResourceAttr("github_codespaces_organization_secret.encrypted_secret", "encrypted_value", secretValueUpdated),
						resource.TestCheckResourceAttrSet("github_codespaces_organization_secret.plaintext_secret", "created_at"),
						resource.TestCheckResourceAttrSet("github_codespaces_organization_secret.plaintext_secret", "updated_at"),
					),
				},
			},
		})
	})

	t.Run("deletes secrets without error", func(t *testing.T) {
		config := `
		resource "github_codespaces_organization_secret" "plaintext_secret" {
			secret_name      = "test_plaintext_secret"
			visibility       = "private"
		}

		resource "github_codespaces_organization_secret" "encrypted_secret" {
			secret_name      = "test_encrypted_secret"
			visibility       = "private"
		}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
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
		resource "github_codespaces_organization_secret" "test_secret" {
			secret_name      = "test_plaintext_secret"
			plaintext_value  = "%s"
			visibility       = "private"
		}
		`, secretValue)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasPaidOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("github_codespaces_organization_secret.test_secret", "plaintext_value", secretValue),
					),
				},
			},
		})
	})
}
