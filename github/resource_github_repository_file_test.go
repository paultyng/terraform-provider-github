package github

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func GenerateKeyMaterial(t *testing.T) (string, string) {
	key, _ := crypto.GenerateKey(t.Name(), "foo@bar.com", "rsa", 2048)
	passphrase := "test_pass"
	key, _ = key.Lock([]byte(passphrase))
	armoredPrivateKey, _ := key.Armor()
	return armoredPrivateKey, passphrase
}

func TestAccGithubRepositoryFile(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	t.Run("creates and manages files", func(t *testing.T) {

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_repository_file" "test" {
				repository      = github_repository.test.name
				branch          = "main"
				file            = "test"
				content         = "bar"
				commit_message  = "Managed by Terraform"
				commit_author   = "Terraform User"
				commit_email    = "terraform@example.com"
				use_contents_api = true
				pgp_signing_key = null
				pgp_signing_key_passphrase = null

			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "content",
				"bar",
			),
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "sha",
				"ba0e162e1c47469e3fe4b393a8bf8c569f302116",
			),
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "ref",
				"main",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_author",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_email",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_message",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_sha",
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
					{
						Config: strings.Replace(config,
							"use_contents_api = true",
							"use_contents_api = false", 1),
						Check: check,
					},
					{
						Config: func() string {
							config := strings.Replace(config,
								"use_contents_api = true",
								"use_contents_api = false", 1,
							)

							armoredPgpKey, PgpPass := GenerateKeyMaterial(t)
							config = strings.Replace(config,
								"pgp_signing_key = null",
								fmt.Sprintf("pgp_signing_key = <<EOT\n%s\nEOT", armoredPgpKey), 1,
							)

							config = strings.Replace(config,
								"pgp_signing_key_passphrase = null",
								fmt.Sprintf("pgp_signing_key_passphrase = \"%s\"", PgpPass), 1,
							)

							return config
						}(),
						Check: check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("can be configured to overwrite files on create", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name      = "tf-acc-test-%s"
			  auto_init = true
			}

			resource "github_repository_file" "test" {
				repository          = github_repository.test.name
				branch              = "main"
				file                = "README.md"
				content             = "overwritten"
				overwrite_on_create = false
			}

		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "content",
				"overwritten",
			),
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "sha",
				"67c1a95c2d9bb138aefeaebb319cca82e531736b",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_author",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_email",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_message",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_sha",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: regexp.MustCompile(`refusing to overwrite existing file`),
					},
					{
						Config: strings.Replace(config,
							"overwrite_on_create = false",
							"overwrite_on_create = true", 1),
						Check: check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("creates and manages files on default branch if branch is omitted", func(t *testing.T) {

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_branch" "test" {
				repository = github_repository.test.name
				branch     = "test"
			}

			resource "github_branch_default" "default"{
				repository = github_repository.test.name
				branch     = github_branch.test.branch
			}

			resource "github_repository_file" "test" {
				depends_on  = [github_branch_default.default]

				repository     = github_repository.test.name
				file           = "test"
				content        = "bar"
				commit_message = "Managed by Terraform"
				commit_author  = "Terraform User"
				commit_email   = "terraform@example.com"
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "content",
				"bar",
			),
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "sha",
				"ba0e162e1c47469e3fe4b393a8bf8c569f302116",
			),
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "ref",
				"test",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_author",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_email",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_message",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_sha",
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

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})

	t.Run("creates and manages files on auto created branch if branch does not exist", func(t *testing.T) {

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_repository_file" "test" {
				repository        = github_repository.test.name
				branch            = "does/not/exist"
				file              = "test"
				content           = "bar"
				commit_message    = "Managed by Terraform"
				commit_author     = "Terraform User"
				commit_email      = "terraform@example.com"
				autocreate_branch = false
			}
		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "content",
				"bar",
			),
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "sha",
				"ba0e162e1c47469e3fe4b393a8bf8c569f302116",
			),
			resource.TestCheckResourceAttr(
				"github_repository_file.test", "ref",
				"does/not/exist",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_author",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_email",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_message",
			),
			resource.TestCheckResourceAttrSet(
				"github_repository_file.test", "commit_sha",
			),
		)

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: regexp.MustCompile(`unexpected status code: 404 Not Found`),
					},
					{
						Config: strings.Replace(config,
							"autocreate_branch = false",
							"autocreate_branch = true", 1),
						Check: check,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})

	})
}
