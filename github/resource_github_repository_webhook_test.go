package github

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubRepositoryWebhook(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	ctx := context.Background()

	githubProvider := Provider()
	diag := githubProvider.Configure(ctx, &terraform.ResourceConfig{})
	if diag.HasError() {
		t.Fatal("err: encountered error while configuring provider")
	}

	var providerOwner string = githubProvider.Meta().(*Owner).name

	t.Run("creates repository webhooks without error", func(t *testing.T) {
		configs := map[string]string{
			"default": fmt.Sprintf(`
				resource "github_repository" "test" {
					name         = "test-%[1]s"
					description  = "Terraform acceptance tests"
				}

				resource "github_repository_webhook" "test" {
					depends_on = ["github_repository.test"]
					repository = "test-%[1]s"

					configuration {
						url          = "https://google.de/webhook"
						content_type = "json"
						insecure_ssl = true
					}

					events = ["pull_request"]
				}
			`, randomID),
			"withOwner": fmt.Sprintf(`
				resource "github_repository" "test" {
					name         = "test-%[1]s"
					description  = "Terraform acceptance tests"
					owner        = "%[2]s"
				}

				resource "github_repository_webhook" "test" {
					depends_on = ["github_repository.test"]
					repository = "test-%[1]s"
					owner      = "%[2]s"

					configuration {
						url          = "https://google.de/webhook"
						content_type = "json"
						insecure_ssl = true
					}

					events = ["pull_request"]
				}
			`, randomID, providerOwner),
		}

		checks := map[string]resource.TestCheckFunc{
			"default": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "active", "true",
				),
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "events.#", "1",
				),
			),
			"withOwner": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "active", "true",
				),
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "events.#", "1",
				),
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "owner", providerOwner,
				),
			),
		}

		testCase := func(t *testing.T, mode string, caseName string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: configs[caseName],
						Check:  checks[caseName],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual, "default")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization, "default")
		})

		t.Run("with a resource-specific owner", func(t *testing.T) {
			testCase(t, organization, "withOwner")
		})
	})

	t.Run("imports repository webhooks without error", func(t *testing.T) {

		configs := map[string]string{
			"default": fmt.Sprintf(`
				resource "github_repository" "test" {
					name         = "test-%[1]s"
					description  = "Terraform acceptance tests"
				}

				resource "github_repository_webhook" "test" {
					depends_on = ["github_repository.test"]
					repository = "test-%[1]s"

					configuration {
						url          = "https://google.de/webhook"
						content_type = "json"
						insecure_ssl = true
					}

					events = ["pull_request"]
				}
			`, randomID),
			"withOwner": fmt.Sprintf(`
				resource "github_repository" "test" {
					name         = "test-%[1]s"
					description  = "Terraform acceptance tests"
					owner        = "%[2]s"
				}

				resource "github_repository_webhook" "test" {
					depends_on = ["github_repository.test"]
					repository = "test-%[1]s"
					owner      = "%[2]s"

					configuration {
						url          = "https://google.de/webhook"
						content_type = "json"
						insecure_ssl = true
					}

					events = ["pull_request"]
				}
			`, randomID, providerOwner),
		}

		check := resource.ComposeTestCheckFunc()

		importStateIdPrefixes := map[string]string{
			"default":   fmt.Sprintf("test-%s/", randomID),
			"withOwner": fmt.Sprintf("%s/test-%s/", providerOwner, randomID),
		}

		testCase := func(t *testing.T, mode string, caseName string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: configs[caseName],
						Check:  check,
					},
					{
						ResourceName:        "github_repository_webhook.test",
						ImportState:         true,
						ImportStateVerify:   true,
						ImportStateIdPrefix: importStateIdPrefixes[caseName],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual, "default")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization, "default")
		})

		t.Run("with a resource-specific owner", func(t *testing.T) {
			testCase(t, organization, "withOwner")
		})
	})

	t.Run("updates repository webhooks without error", func(t *testing.T) {

		configs := map[string]string{
			"before": fmt.Sprintf(`
				resource "github_repository" "test" {
				  name         = "test-%[1]s"
				  description  = "Terraform acceptance tests"
				}

				resource "github_repository_webhook" "test" {
				  depends_on = ["github_repository.test"]
				  repository = "test-%[1]s"

				  configuration {
				    url          = "https://google.de/webhook"
				    content_type = "json"
				    insecure_ssl = true
				  }

				  events = ["pull_request"]
				}
			`, randomID),
			"after": fmt.Sprintf(`
				resource "github_repository" "test" {
				  name         = "test-%[1]s"
				  description  = "Terraform acceptance tests"
				}

				resource "github_repository_webhook" "test" {
				  depends_on = ["github_repository.test"]
				  repository = "test-%[1]s"

				  configuration {
				    secret       = "secret"
				    url          = "https://google.de/webhook"
				    content_type = "json"
				    insecure_ssl = true
				  }

				  events = ["pull_request"]
				}
			`, randomID),
			"beforeWithOwner": fmt.Sprintf(`
				resource "github_repository" "test" {
				  name         = "test-%[1]s"
				  description  = "Terraform acceptance tests"
					owner        = "%[2]s"
				}

				resource "github_repository_webhook" "test" {
				  depends_on = ["github_repository.test"]
				  repository = "test-%[1]s"
					owner      = "%[2]s"

				  configuration {
				    url          = "https://google.de/webhook"
				    content_type = "json"
				    insecure_ssl = true
				  }

				  events = ["pull_request"]
				}
			`, randomID, providerOwner),
			"afterWithOwner": fmt.Sprintf(`
				resource "github_repository" "test" {
				  name         = "test-%[1]s"
				  description  = "Terraform acceptance tests"
					owner        = "%[2]s"
				}

				resource "github_repository_webhook" "test" {
				  depends_on = ["github_repository.test"]
				  repository = "test-%[1]s"
					owner      = "%[2]s"

				  configuration {
				    secret       = "secret"
				    url          = "https://google.de/webhook"
				    content_type = "json"
				    insecure_ssl = true
				  }

				  events = ["pull_request"]
				}
			`, randomID, providerOwner),
		}

		checks := map[string]resource.TestCheckFunc{
			"before": resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "events.#", "1",
			),
			"after": resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "configuration.0.secret", "secret",
			),
			"beforeWithOwner": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "events.#", "1",
				),
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "owner", providerOwner,
				),
			),
			"afterWithOwner": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "configuration.0.secret", "secret",
				),
				resource.TestCheckResourceAttr(
					"github_repository_webhook.test", "owner", providerOwner,
				),
			),
		}

		testCase := func(t *testing.T, mode string, beforeCaseName string, afterCaseName string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: configs[beforeCaseName],
						Check:  checks[beforeCaseName],
					},
					{
						Config: configs[afterCaseName],
						Check:  checks[afterCaseName],
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual, "before", "after")
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization, "before", "after")
		})

		t.Run("with a resource-specific owner", func(t *testing.T) {
			testCase(t, organization, "beforeWithOwner", "afterWithOwner")
		})
	})

	t.Run("creates repository webhooks with unique owner without error", func(t *testing.T) {
		testResourceOwner := resourceOwner()
		config := fmt.Sprintf(`
				resource "github_repository" "test" {
					name         = "test-%[1]s"
					description  = "Terraform acceptance tests"
					owner        = "%[2]s"
				}

				resource "github_repository_webhook" "test" {
					depends_on = ["github_repository.test"]
					repository = "test-%[1]s"
					owner      = "%[2]s"

					configuration {
						url          = "https://google.de/webhook"
						content_type = "json"
						insecure_ssl = true
					}

					events = ["pull_request"]
				}
			`, randomID, testResourceOwner)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "active", "true",
			),
			resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "events.#", "1",
			),
			resource.TestCheckResourceAttr(
				"github_repository_webhook.test", "owner", testResourceOwner,
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

		t.Run("with a unique resource-specific owner", func(t *testing.T) {
			if testResourceOwner == "" {
				t.Skipf("Skipping %s which requires GITHUB_RESOURCE_OWNER to be set to a non-empty value", t.Name())
			}
			if testResourceOwner == providerOwner {
				t.Fatalf("err: webhook owner %[1]s was the same as provider owner %[2]s; they must be different for this test (hint: set GITHUB_RESOURCE_OWNER to be different than provider owner '%[2]s', or set the provider owner to something different)", testResourceOwner, providerOwner)
			}
			testCase(t, enterprise)
		})
	})
}
