package github

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGithubOrganizationRole(t *testing.T) {
	t.Run("get the organization role without error", func(t *testing.T) {
		roleId := 138
		config := fmt.Sprintf(`
			data "github_organization_role" "test" {
				role_id = %d
			}
		`, roleId)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.github_organization_role.test", "role_id", strconv.Itoa(roleId)),
						resource.TestCheckResourceAttr("data.github_organization_role.test", "name", "security_manager"),
						resource.TestCheckResourceAttr("data.github_organization_role.test", "source", "Predefined"),
						resource.TestCheckResourceAttr("data.github_organization_role.test", "base_role", "read"),
						resource.TestCheckResourceAttrSet("data.github_organization_role.test", "description"),
						resource.TestCheckResourceAttrSet("data.github_organization_role.test", "permissions.#"),
					),
				},
			},
		})
	})
}
