package github

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccOrganizationBlock_basic(t *testing.T) {
	t.Run("creates organization block", func(t *testing.T) {
		config := `
resource "github_organization_block" "test" {
  username = "cgriggs01"
}
`

		rn := "github_organization_block.test"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccOrganizationBlockDestroy,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckOrganizationBlockExists(rn),
					),
				},
				{
					ResourceName:      rn,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func testAccOrganizationBlockDestroy(s *terraform.State) error {
	meta, err := getTestMeta()
	if err != nil {
		return err
	}
	conn := meta.v3client
	orgName := meta.name

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_organization_block" {
			continue
		}

		username := rs.Primary.ID

		res, err := conn.Organizations.UnblockUser(context.TODO(), orgName, username)
		if res.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckOrganizationBlockExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		username := rs.Primary.ID
		meta, err := getTestMeta()
		if err != nil {
			return err
		}
		conn := meta.v3client
		orgName := meta.name

		blocked, _, err := conn.Organizations.IsBlocked(context.TODO(), orgName, username)
		if err != nil {
			return err
		}
		if !blocked {
			return fmt.Errorf("not blocked: %s %s", orgName, username)
		}
		return nil
	}
}
