package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryCustomProperty(t *testing.T) {

	t.Skip("You need an org with custom properties already setup as described in the variables below") // TODO: at the time of writing org_custom_properties are not supported by this terraform provider, so cant be setup in the test itself for now
	singleSelectPropertyName := "single-select"                                                        // Needs to be a of type single_select, and have "option1" as an option
	multiSelectPropertyName := "multi-select"                                                          // Needs to be a of type multi_select, and have "option1" and "option2" as an options
	trueFlasePropertyName := "true-false"                                                              // Needs to be a of type true_false
	stringPropertyName := "string"                                                                     // Needs to be a of type string

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates custom property of type single_select without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = true
			}
			resource "github_repository_custom_property" "test" {
				repository    = github_repository.test.name
				property_name = "%s"
				property_type = "single_select"
				property_value = ["option1"]
			}
		`, randomID, singleSelectPropertyName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_name", singleSelectPropertyName),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.#", "1"),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.0", "option1"),
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

	t.Run("creates custom property of type multi_select without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = true
			}
			resource "github_repository_custom_property" "test" {
				repository    = github_repository.test.name
				property_name = "%s"
				property_type = "multi_select"
				property_value = ["option1", "option2"]
			}
		`, randomID, multiSelectPropertyName)

		checkWithOwner := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_name", multiSelectPropertyName),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.#", "2"),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.0", "option1"),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.1", "option2"),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checkWithOwner,
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

	t.Run("creates custom property of type true-false without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = true
			}
			resource "github_repository_custom_property" "test" {
				repository    = github_repository.test.name
				property_name = "%s"
				property_type = "true_false"
				property_value = ["true"]
			}
		`, randomID, trueFlasePropertyName)

		checkWithOwner := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_name", trueFlasePropertyName),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.#", "1"),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.0", "true"),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checkWithOwner,
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

	t.Run("creates custom property of type string without error", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = true
			}
			resource "github_repository_custom_property" "test" {
				repository    = github_repository.test.name
				property_name = "%s"
				property_type = "string"
				property_value = ["text"]
			}
		`, randomID, stringPropertyName)

		checkWithOwner := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_name", stringPropertyName),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.#", "1"),
			resource.TestCheckResourceAttr("github_repository_custom_property.test", "property_value.0", "text"),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  checkWithOwner,
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
