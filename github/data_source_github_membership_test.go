package github

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubMembershipDataSource(t *testing.T) {
	if len(testAccConf.testOrgUser) == 0 {
		t.Skip("No org user provided")
	}

	t.Run("queries the membership for a user in a specified organization", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_membership" "test" {
				username = "%s"
				organization = "%s"
			}
		`, testAccConf.testOrgUser, testAccConf.owner)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.github_membership.test", "username", testAccConf.testOrgUser),
			resource.TestCheckResourceAttrSet("data.github_membership.test", "role"),
			resource.TestCheckResourceAttrSet("data.github_membership.test", "etag"),
			resource.TestCheckResourceAttrSet("data.github_membership.test", "state"),
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

	t.Run("errors when querying with non-existent user", func(t *testing.T) {
		config := fmt.Sprintf(`
			data "github_membership" "test" {
				username = "!%s"
				organization = "%s"
			}
		`, testAccConf.testOrgUser, testAccConf.owner)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(`Not Found`),
				},
			},
		})
	})
}
