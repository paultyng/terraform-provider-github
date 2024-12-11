package github

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubEnterpriseOrganization(t *testing.T) {
	t.Run("creates and updates an enterprise organization without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-%s", randomID)

		desc := "Initial org description"
		updatedDesc := "Updated org description"

		config := fmt.Sprintf(`
		  data "github_enterprise" "enterprise" {
			slug = "%s"
		  }

		  data "github_user" "current" {
			username = ""
		  }

		  resource "github_enterprise_organization" "org" {
			enterprise_id = data.github_enterprise.enterprise.id
			name          = "%s"
			description   = "%s"
			billing_email = data.github_user.current.email
			admin_logins  = [
			  data.github_user.current.login
			]
		  }
			`, testAccConf.enterpriseSlug, orgName, desc)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet(
					"github_enterprise_organization.org", "enterprise_id",
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "name",
					orgName,
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					desc,
				),
				resource.TestCheckResourceAttrSet(
					"github_enterprise_organization.org", "billing_email",
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "admin_logins.#",
					"1",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					updatedDesc,
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks["before"],
				},
				{
					Config: strings.Replace(config,
						desc,
						updatedDesc, 1),
					Check: checks["after"],
				},
			},
		})
	})

	t.Run("deletes an enterprise organization without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-%s", randomID)

		config := fmt.Sprintf(`
		  data "github_enterprise" "enterprise" {
			slug = "%s"
		  }

		  data "github_user" "current" {
			username = ""
		  }

		  resource "github_enterprise_organization" "org" {
			enterprise_id = data.github_enterprise.enterprise.id
			name          = "%s"
			billing_email = data.github_user.current.email
			admin_logins  = [
			  data.github_user.current.login
			]
		  }
			`, testAccConf.enterpriseSlug, orgName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:  config,
					Destroy: true,
				},
			},
		})
	})

	t.Run("creates and updates org with display name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-displayname%s", randomID)

		displayName := fmt.Sprintf("Tf Acc Test displayname %s", randomID)
		updatedDisplayName := fmt.Sprintf("Updated Tf Acc Test Display Name %s", randomID)

		desc := "Initial org description"
		updatedDesc := "Updated org description"

		config := fmt.Sprintf(`
		  data "github_enterprise" "enterprise" {
			slug = "%s"
		  }

		  data "github_user" "current" {
			username = ""
		  }

		  resource "github_enterprise_organization" "org" {
			enterprise_id = data.github_enterprise.enterprise.id
			name          = "%s"
			display_name  = "%s"
			description   = "%s"
			billing_email = data.github_user.current.email
			admin_logins  = [
			  data.github_user.current.login
			]
		  }
			`, testAccConf.enterpriseSlug, orgName, displayName, desc)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet(
					"github_enterprise_organization.org", "enterprise_id",
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "name",
					orgName,
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "display_name",
					displayName,
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					desc,
				),
				resource.TestCheckResourceAttrSet(
					"github_enterprise_organization.org", "billing_email",
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "admin_logins.#",
					"1",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					updatedDesc,
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "display_name",
					updatedDisplayName,
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks["before"],
				},
				{
					Config: strings.Replace(
						strings.Replace(config,
							displayName,
							updatedDisplayName, 1),
						desc,
						updatedDesc, 1),
					Check: checks["after"],
				},
			},
		})
	})

	t.Run("creates org without display name, set and update display name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-adddisplayname%s", randomID)

		displayName := fmt.Sprintf("Tf Acc Test Add displayname %s", randomID)
		updatedDisplayName := fmt.Sprintf("Updated Tf Acc Test Add Display Name %s", randomID)

		desc := "Initial org description"
		updatedDesc := "Updated org description"

		configWithoutDisplayName := fmt.Sprintf(`
		  data "github_enterprise" "enterprise" {
			slug = "%s"
		  }

		  data "github_user" "current" {
			username = ""
		  }

		  resource "github_enterprise_organization" "org" {
			enterprise_id = data.github_enterprise.enterprise.id
			name          = "%s"
			description   = "%s"
			billing_email = data.github_user.current.email
			admin_logins  = [
			  data.github_user.current.login
			]
		  }
			`, testAccConf.enterpriseSlug, orgName, desc)

		configWithDisplayName := fmt.Sprintf(`
			data "github_enterprise" "enterprise" {
			  slug = "%s"
			}

			data "github_user" "current" {
			  username = ""
			}

			resource "github_enterprise_organization" "org" {
			  enterprise_id = data.github_enterprise.enterprise.id
			  name          = "%s"
			  display_name  = "%s"
			  description   = "%s"
			  billing_email = data.github_user.current.email
			  admin_logins  = [
				data.github_user.current.login
			  ]
			}
			  `, testAccConf.enterpriseSlug, orgName, displayName, desc)

		checks := map[string]resource.TestCheckFunc{
			"create": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet(
					"github_enterprise_organization.org", "enterprise_id",
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "name",
					orgName,
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					desc,
				),
				resource.TestCheckResourceAttrSet(
					"github_enterprise_organization.org", "billing_email",
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "admin_logins.#",
					"1",
				),
			),
			"set": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					desc,
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "display_name",
					displayName,
				),
			),
			"updateDisplayName": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					desc,
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "display_name",
					updatedDisplayName,
				),
			),
			"updateDesc": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					updatedDesc,
				),
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "display_name",
					updatedDisplayName,
				),
			),
			"unset": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_enterprise_organization.org", "description",
					desc,
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: configWithoutDisplayName,
					Check:  checks["create"],
				},
				{
					Config: configWithDisplayName,
					Check:  checks["set"],
				},
				{
					Config: strings.Replace(configWithDisplayName,
						displayName,
						updatedDisplayName, 1),
					Check: checks["updateDisplayName"],
				},
				{
					Config: strings.Replace(
						strings.Replace(configWithDisplayName,
							displayName,
							updatedDisplayName, 1),
						desc,
						updatedDesc, 1),
					Check: checks["updateDesc"],
				},
				{
					Config: configWithoutDisplayName,
					Check:  checks["unset"],
				},
			},
		})
	})

	t.Run("imports enterprise organization without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-import%s", randomID)

		config := fmt.Sprintf(`
			data "github_enterprise" "enterprise" {
			  slug = "%s"
			}

			data "github_user" "current" {
			  username = ""
			}

			resource "github_enterprise_organization" "org" {
			  enterprise_id = data.github_enterprise.enterprise.id
			  name          = "%s"
			  billing_email = data.github_user.current.email
			  admin_logins  = [
				data.github_user.current.login
			  ]
			}
			  `, testAccConf.enterpriseSlug, orgName)

		check := resource.ComposeTestCheckFunc()

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:      "github_enterprise_organization.org",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateId:     fmt.Sprintf(`%s/%s`, testAccConf.enterpriseSlug, orgName),
				},
			},
		})
	})

	t.Run("imports enterprise organization invalid enterprise name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-adddisplayname%s", randomID)

		config := fmt.Sprintf(`
			data "github_enterprise" "enterprise" {
			  slug = "%s"
			}

			data "github_user" "current" {
			  username = ""
			}

			resource "github_enterprise_organization" "org" {
			  enterprise_id = data.github_enterprise.enterprise.id
			  name          = "%s"
			  description   = "org description"
			  billing_email = data.github_user.current.email
			  admin_logins  = [
				data.github_user.current.login
			  ]
			}
			  `, testAccConf.enterpriseSlug, orgName)

		check := resource.ComposeTestCheckFunc()

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:  "github_enterprise_organization.org",
					ImportState:   true,
					ImportStateId: fmt.Sprintf(`%s/%s`, randomID, orgName),
					ExpectError:   regexp.MustCompile("Could not resolve to a Business with the URL slug of .*"),
				},
			},
		})
	})

	t.Run("imports enterprise organization invalid organization name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		orgName := fmt.Sprintf("tf-acc-test-adddisplayname%s", randomID)

		config := fmt.Sprintf(`
			data "github_enterprise" "enterprise" {
			  slug = "%s"
			}

			data "github_user" "current" {
			  username = ""
			}

			resource "github_enterprise_organization" "org" {
			  enterprise_id = data.github_enterprise.enterprise.id
			  name          = "%s"
			  description   = "org description"
			  billing_email = data.github_user.current.email
			  admin_logins  = [
				data.github_user.current.login
			  ]
			}
			  `, testAccConf.enterpriseSlug, orgName)

		check := resource.ComposeTestCheckFunc()

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessMode(t, enterprise) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:  "github_enterprise_organization.org",
					ImportState:   true,
					ImportStateId: fmt.Sprintf(`%s/%s`, testAccConf.enterpriseSlug, randomID),
					ExpectError:   regexp.MustCompile("Could not resolve to an Organization with the login of .*"),
				},
			},
		})
	})
}
