package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsRepositoryOIDCSubjectClaimCustomizationTemplate(t *testing.T) {
	t.Run("creates repository oidc subject claim customization template without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
		resource "github_repository" "test" {
			name = "tf-acc-test-%s"
			visibility = "private"
		}

		resource "github_actions_repository_oidc_subject_claim_customization_template" "test" {
			repository = github_repository.test.name
			use_default = false
			include_claim_keys = ["repo", "context", "job_workflow_ref"]
		}`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"use_default", "false"),
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"include_claim_keys.#", "3",
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"include_claim_keys.0", "repo",
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"include_claim_keys.1", "context",
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"include_claim_keys.2", "job_workflow_ref",
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

	t.Run("updates repository oidc subject claim customization template without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		configTemplate := `
		resource "github_repository" "test" {
			name = "tf-acc-test-%s"
			visibility = "private"
		}

		resource "github_actions_repository_oidc_subject_claim_customization_template" "test" {
			repository = github_repository.test.name
			use_default = %t
			include_claim_keys = %s
		}`

		claims := `["repository_owner_id", "run_id", "workflow"]`
		updatedClaims := `["actor", "actor_id", "head_ref", "repository"]`

		resetToDefaultConfigTemplate := `
		resource "github_repository" "test" {
			name = "tf-acc-test-%s"
			visibility = "private"
		}

		resource "github_actions_repository_oidc_subject_claim_customization_template" "test" {
			repository = github_repository.test.name
			use_default = true
		}
`

		configs := map[string]string{
			"before": fmt.Sprintf(configTemplate, randomID, false, claims),

			"after": fmt.Sprintf(configTemplate, randomID, false, updatedClaims),

			"reset_to_default": fmt.Sprintf(resetToDefaultConfigTemplate, randomID),
		}
		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"use_default", "false"),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.#", "3",
				),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.0", "repository_owner_id",
				),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.1", "run_id",
				),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.2", "workflow",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"use_default", "false"),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.#", "4",
				),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.0", "actor",
				),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.1", "actor_id",
				),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.2", "head_ref",
				),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.3", "repository",
				),
			),
			"reset_to_default": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"use_default", "true"),
				resource.TestCheckResourceAttr(
					"github_actions_repository_oidc_subject_claim_customization_template.test",
					"include_claim_keys.#", "0",
				),
			),
		}
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: configs["before"],
					Check:  checks["before"],
				},
				{
					Config: configs["after"],
					Check:  checks["after"],
				},
				{
					Config: configs["reset_to_default"],
					Check:  checks["reset_to_default"],
				},
			},
		})
	})

	t.Run("imports repository oidc subject claim customization template without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
		resource "github_repository" "test" {
			name = "tf-acc-test-%s"
			visibility = "private"
		}
		resource "github_actions_repository_oidc_subject_claim_customization_template" "test" {
			repository = github_repository.test.name
			use_default = false
			include_claim_keys = ["repository_owner_id", "run_id", "workflow"]
		}`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"use_default", "false",
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"include_claim_keys.#", "3",
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"include_claim_keys.0", "repository_owner_id",
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"include_claim_keys.1", "run_id",
			),
			resource.TestCheckResourceAttr(
				"github_actions_repository_oidc_subject_claim_customization_template.test",
				"include_claim_keys.2", "workflow",
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
				{
					ResourceName:      "github_actions_repository_oidc_subject_claim_customization_template.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
