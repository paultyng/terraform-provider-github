package github

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsEnvironmentSecret(t *testing.T) {
	t.Run("creates and updates secrets without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		secretValue := base64.StdEncoding.EncodeToString([]byte("super_secret_value"))
		updatedSecretValue := base64.StdEncoding.EncodeToString([]byte("updated_super_secret_value"))

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "tf-acc-test-%s"
			}

			resource "github_repository_environment" "test" {
			  repository       = github_repository.test.name
			  environment      = "environment / test"
			}

			resource "github_actions_environment_secret" "plaintext_secret" {
			  repository       = github_repository.test.name
			  environment      = github_repository_environment.test.environment
			  secret_name      = "test_plaintext_secret_name"
			  plaintext_value  = "%s"
			}

			resource "github_actions_environment_secret" "encrypted_secret" {
			  repository       = github_repository.test.name
			  environment      = github_repository_environment.test.environment
			  secret_name      = "test_encrypted_secret_name"
			  encrypted_value  = "%s"
			}
		`, randomID, secretValue, secretValue)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_environment_secret.plaintext_secret", "plaintext_value",
					secretValue,
				),
				resource.TestCheckResourceAttr(
					"github_actions_environment_secret.encrypted_secret", "encrypted_value",
					secretValue,
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_environment_secret.plaintext_secret", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_environment_secret.plaintext_secret", "updated_at",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_environment_secret.plaintext_secret", "plaintext_value",
					updatedSecretValue,
				),
				resource.TestCheckResourceAttr(
					"github_actions_environment_secret.encrypted_secret", "encrypted_value",
					updatedSecretValue,
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_environment_secret.plaintext_secret", "created_at",
				),
				resource.TestCheckResourceAttrSet(
					"github_actions_environment_secret.plaintext_secret", "updated_at",
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
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
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		secretValue := base64.StdEncoding.EncodeToString([]byte("super_secret_value"))

		config := fmt.Sprintf(`
				resource "github_repository" "test" {
					name = "tf-acc-test-%s"
				}

				resource "github_repository_environment" "test" {
					repository       = github_repository.test.name
					environment      = "environment / test"
				}

				resource "github_actions_environment_secret" "plaintext_secret" {
					repository       = github_repository.test.name
					environment      = github_repository_environment.test.environment
					secret_name      = "test_plaintext_secret_name"
					plaintext_value  = "%s"
				}

				resource "github_actions_environment_secret" "encrypted_secret" {
					repository       = github_repository.test.name
					environment      = github_repository_environment.test.environment
					secret_name      = "test_encrypted_secret_name"
					encrypted_value  = "%s"
				}
			`, randomID, secretValue, secretValue)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:  config,
					Destroy: true,
				},
			},
		})
	})
}
