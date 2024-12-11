package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubEnterpriseDataSource(t *testing.T) {
	config := fmt.Sprintf(`
			data "github_enterprise" "test" {
				slug = "%s"
			}
		`,
		testAccConf.enterpriseSlug,
	)

	check := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.github_enterprise.test", "slug", testAccConf.enterpriseSlug),
		resource.TestCheckResourceAttrSet("data.github_enterprise.test", "name"),
		resource.TestCheckResourceAttrSet("data.github_enterprise.test", "created_at"),
		resource.TestCheckResourceAttrSet("data.github_enterprise.test", "url"),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { skipUnlessMode(t, enterprise) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	},
	)
}
