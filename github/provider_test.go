package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	t.Run("runs internal validation without error", func(t *testing.T) {
		if err := Provider().InternalValidate(); err != nil {
			t.Fatalf("err: %s", err)
		}
	})

	t.Run("has an implementation", func(t *testing.T) {
		// FIXME: unsure if this is useful; refactored from:
		// func TestProvider_impl(t *testing.T) {
		// 	var _ terraform.ResourceProvider = Provider()
		// }

		var _ schema.Provider = *Provider()
	})
}

func TestAccProviderConfigure(t *testing.T) {
	t.Run("can be configured to run anonymously", func(t *testing.T) {
		config := `
		provider "github" {
		  token = ""
		}
		data "github_ip_ranges" "test" {}
		`

		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:             config,
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("can be configured to run insecurely", func(t *testing.T) {
		config := `
		provider "github" {
			token = ""
			insecure = true
		}
		data "github_ip_ranges" "test" {}
    `

		resource.Test(t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:             config,
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("can be configured with an individual account", func(t *testing.T) {
		config := fmt.Sprintf(`
			provider "github" {
				token = "%s"
				owner = "%s"
			}
			`, testAccConf.token, testAccConf.owner)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, individual) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:             config,
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("can be configured with an organization account", func(t *testing.T) {
		config := fmt.Sprintf(`
			provider "github" {
				token = "%s"
				organization = "%s"
			}`, testAccConf.token, testAccConf.owner)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, organization) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:             config,
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("can be configured with a GHES deployment", func(t *testing.T) {
		config := fmt.Sprintf(`
			provider "github" {
				token = "%s"
				base_url = "%s"
			}`, testAccConf.token, testAccConf.owner)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, individual) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:             config,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("can be configured with max retries", func(t *testing.T) {
		config := fmt.Sprintf(`
			provider "github" {
				token = "%s"
				owner = "%s"
				max_retries = 3
			}`, testAccConf.token, testAccConf.owner)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, individual) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:             config,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}
