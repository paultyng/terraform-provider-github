package github

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubOrganizationTeamSyncGroupsDataSource_existing(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { skipUnlessMode(t, enterprise) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "github_organization_team_sync_groups" "test" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.github_organization_team_sync_groups.test", "groups.#"),
					resource.TestCheckResourceAttrSet("data.github_organization_team_sync_groups.test", "groups.0.group_id"),
					resource.TestCheckResourceAttrSet("data.github_organization_team_sync_groups.test", "groups.0.group_name"),
				),
			},
		},
	})
}
