package github

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestGithubIPAllowListEntry(t *testing.T) {
	if isEnterprise != "true" {
		t.Skip("Skipping because `ENTERPRISE_ACCOUNT` is not set or set to false")
	}

	if testEnterprise == "" {
		t.Skip("Skipping because `ENTERPRISE_SLUG` is not set")
	}

	randomName := acctest.RandStringFromCharSet(20, acctest.CharSetAlphaNum)

	t.Run("Creates an IP allow list entry without errors", func(t *testing.T) {

		config := fmt.Sprintf(`
			data "github_organization" "test" {
				name         = "%s"
				summary_only = true
			}

			resource "github_ip_allow_list_entry" "test" {
				value  = "127.0.0.1"
				active = false
				name   = "%s"
				owner  = data.github_organization.test.node_id
			}
		`, testEnterprise, randomName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_ip_allow_list_entry.test", "name",
				randomName,
			),
			resource.TestCheckResourceAttr(
				"github_ip_allow_list_entry.test", "active",
				"false",
			),
			resource.TestCheckResourceAttr(
				"github_ip_allow_list_entry.test", "value",
				"127.0.0.1",
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
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Updates an ip allow list entry without error", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_organization" "test" {
				name         = "%s"
				summary_only = true
			}

			resource "github_ip_allow_list_entry" "test" {
				value  = "127.0.0.1"
				active = false
				name   = "%s"
				owner  = data.github_organization.test.node_id
			}
		`, testEnterprise, randomName)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_ip_allow_list_entry.test", "name",
					randomName,
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_ip_allow_list_entry.test", "name",
					randomName+"renamed",
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
						// Rename the entry to something else
						Config: strings.Replace(
							config,
							randomName,
							randomName+"renamed", 1),
						Check: checks["after"],
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

	t.Run("Imports an ip allow list entry without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			data "github_organization" "test" {
				name         = "%s"
				summary_only = true
			}

			resource "github_ip_allow_list_entry" "test" {
				value  = "127.0.0.1"
				active = false
				name   = "%s"
				owner  = data.github_organization.test.node_id
			}
		`, testEnterprise, randomName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_ip_allow_list_entry.test", "name"),
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
						ResourceName:      "github_ip_allow_list_entry.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		}

		t.Run("with an enterprise account", func(t *testing.T) {
			testCase(t, enterprise)
		})

	})

}
