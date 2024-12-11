package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplateDataSource(t *testing.T) {
	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("get an repository oidc subject claim customization template without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				visibility = "private"
			}

			resource "github_actions_repository_oidc_subject_claim_customization_template" "test" {
				repository = github_repository.test.name
				use_default = false
				include_claim_keys = ["repo", "context", "job_workflow_ref"]
			}
		`, randomID)

		config2 := config + `
			data "github_actions_repository_oidc_subject_claim_customization_template" "test" {
				name = github_repository.test.name
			}
		`

		config3 := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				visibility = "private"
			}

			resource "github_actions_repository_oidc_subject_claim_customization_template" "test" {
				repository = github_repository.test.name
				use_default = true
			}
		`, randomID)

		config4 := config3 + `
			data "github_actions_repository_oidc_subject_claim_customization_template" "test" {
				name = github_repository.test.name
			}
		`

		check1 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_actions_repository_oidc_subject_claim_customization_template.test", "use_default", "false"),
			resource.TestCheckResourceAttr("data.github_actions_repository_oidc_subject_claim_customization_template.test", "include_claim_keys.#", "3"),
			resource.TestCheckResourceAttr("data.github_actions_repository_oidc_subject_claim_customization_template.test", "include_claim_keys.0", "repo"),
			resource.TestCheckResourceAttr("data.github_actions_repository_oidc_subject_claim_customization_template.test", "include_claim_keys.1", "context"),
			resource.TestCheckResourceAttr("data.github_actions_repository_oidc_subject_claim_customization_template.test", "include_claim_keys.2", "job_workflow_ref"),
		)

		check2 := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_actions_repository_oidc_subject_claim_customization_template.test", "use_default", "true"),
			resource.TestCheckResourceAttr("data.github_actions_repository_oidc_subject_claim_customization_template.test", "include_claim_keys.#", "0"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  resource.ComposeTestCheckFunc(),
				},
				{
					Config: config2,
					Check:  check1,
				},
				{
					Config: config3,
					Check:  resource.ComposeTestCheckFunc(),
				},
				{
					Config: config4,
					Check:  check2,
				},
			},
		})
	})
}
