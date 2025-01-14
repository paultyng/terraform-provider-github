package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryCustomPropertiesDataSource(t *testing.T) {

	t.Skip("You need an org with custom properties already setup as described in the variables below") // TODO: at the time of writing org_custom_properties are not supported by this terraform provider, so cant be setup in the test itself for now
	singleSelectPropertyName := "single-select"                                                        // Needs to be a of type single_select, and have "option1" as an option
	multiSelectPropertyName := "multi-select"                                                          // Needs to be a of type multi_select, and have "option1" and "option2" as an options
	trueFlasePropertyName := "true-false"                                                              // Needs to be a of type true_false, and have "option1" as an option
	stringPropertyName := "string"                                                                     // Needs to be a of type string, and have "option1" as an option

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
			data "github_repository_custom_properties" "test" {
				repository    = github_repository_custom_property.test.repository
			}
		`, randomID, singleSelectPropertyName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckTypeSetElemNestedAttrs("data.github_repository_custom_properties.test",
				"property.*", map[string]string{
					"property_name":    singleSelectPropertyName,
					"property_value.#": "1",
					"property_value.0": "option1",
				}),
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
			data "github_repository_custom_properties" "test" {
				repository    = github_repository_custom_property.test.repository
			}
		`, randomID, multiSelectPropertyName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckTypeSetElemNestedAttrs("data.github_repository_custom_properties.test",
				"property.*", map[string]string{
					"property_name":    multiSelectPropertyName,
					"property_value.#": "2",
					"property_value.0": "option1",
					"property_value.1": "option2",
				}),
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

	t.Run("creates custom property of type true_false without error", func(t *testing.T) {

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
			data "github_repository_custom_properties" "test" {
				repository    = github_repository_custom_property.test.repository
			}
		`, randomID, trueFlasePropertyName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckTypeSetElemNestedAttrs("data.github_repository_custom_properties.test",
				"property.*", map[string]string{
					"property_name":    trueFlasePropertyName,
					"property_value.#": "1",
					"property_value.0": "true",
				}),
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
			data "github_repository_custom_properties" "test" {
				repository    = github_repository_custom_property.test.repository
			}
		`, randomID, stringPropertyName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckTypeSetElemNestedAttrs("data.github_repository_custom_properties.test",
				"property.*", map[string]string{
					"property_name":    stringPropertyName,
					"property_value.#": "1",
					"property_value.0": "text",
				}),
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
}
