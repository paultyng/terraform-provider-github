package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubActionsOrganizationOIDCSubjectClaimCustomizationTemplate(t *testing.T) {
	t.Run("creates organization oidc subject claim customization template without error", func(t *testing.T) {
		config := `
		resource "github_actions_organization_oidc_subject_claim_customization_template" "test" {
			include_claim_keys = ["repo", "context", "job_workflow_ref"]
		}`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_oidc_subject_claim_customization_template.test",
				"include_claim_keys.#", "3",
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_oidc_subject_claim_customization_template.test",
				"include_claim_keys.0", "repo",
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_oidc_subject_claim_customization_template.test",
				"include_claim_keys.1", "context",
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_oidc_subject_claim_customization_template.test",
				"include_claim_keys.2", "job_workflow_ref",
			),
		)
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})

	t.Run("updates organization oidc subject claim customization template without error", func(t *testing.T) {
		resourceTemplate := `
		resource "github_actions_organization_oidc_subject_claim_customization_template" "test" {
			include_claim_keys = %s
		}`

		claims := `["repository_owner_id", "run_id", "workflow"]`
		updatedClaims := `["actor", "actor_id", "head_ref", "repository"]`

		configs := map[string]string{
			"before": fmt.Sprintf(resourceTemplate, claims),

			"after": fmt.Sprintf(resourceTemplate, updatedClaims),
		}
		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.#", "3",
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.0", "repository_owner_id",
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.1", "run_id",
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.2", "workflow",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.#", "4",
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.0", "actor",
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.1", "actor_id",
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.2", "head_ref",
				),
				resource.TestCheckResourceAttr(
					"github_actions_organization_oidc_subject_claim_customization_template.test",
					"include_claim_keys.3", "repository",
				),
			),
		}
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
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
			},
		})
	})

	t.Run("imports organization oidc subject claim customization template without error", func(t *testing.T) {
		config := `
		resource "github_actions_organization_oidc_subject_claim_customization_template" "test" {
			include_claim_keys = ["repository_owner_id", "run_id", "workflow"]
		}`

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_actions_organization_oidc_subject_claim_customization_template.test",
				"include_claim_keys.#", "3",
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_oidc_subject_claim_customization_template.test",
				"include_claim_keys.0", "repository_owner_id",
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_oidc_subject_claim_customization_template.test",
				"include_claim_keys.1", "run_id",
			),
			resource.TestCheckResourceAttr(
				"github_actions_organization_oidc_subject_claim_customization_template.test",
				"include_claim_keys.2", "workflow",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:      "github_actions_organization_oidc_subject_claim_customization_template.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}
