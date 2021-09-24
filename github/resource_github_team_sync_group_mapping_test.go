package github

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v39/github"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccGithubTeamSyncGroupMapping_basic(t *testing.T) {
	if isEnterprise != "true" {
		t.Skip("Skipping because `ENTERPRISE_ACCOUNT` is not set or set to false")
	}
	teamName := acctest.RandomWithPrefix("tf-acc-test-%s")
	rn := "github_team_sync_group_mapping.test_mapping"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGithubTeamSyncGroupMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubTeamSyncGroupMappingConfig(teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "group.#", "3"),
					resource.TestCheckResourceAttrSet(rn, "group.3924494127.group_id"),
					resource.TestCheckResourceAttrSet(rn, "group.3924494127.group_name"),
					resource.TestCheckResourceAttrSet(rn, "group.4283356133.group_id"),
					resource.TestCheckResourceAttrSet(rn, "group.4283356133.group_name"),
					resource.TestCheckResourceAttrSet(rn, "group.451718421.group_id"),
					resource.TestCheckResourceAttrSet(rn, "group.451718421.group_name"),
				),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccGithubTeamSyncGroupMappingImportStateIdFunc(rn),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGithubTeamSyncGroupMapping_disappears(t *testing.T) {
	if isEnterprise != "true" {
		t.Skip("Skipping because `ENTERPRISE_ACCOUNT` is not set or set to false")
	}
	teamName := acctest.RandomWithPrefix("tf-acc-test-%s")
	rn := "github_team_sync_group_mapping.test_mapping"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGithubTeamSyncGroupMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubTeamSyncGroupMappingConfig(teamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubTeamSyncGroupMappingDisappears(rn),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGithubTeamSyncGroupMapping_update(t *testing.T) {
	if isEnterprise != "true" {
		t.Skip("Skipping because `ENTERPRISE_ACCOUNT` is not set or set to false")
	}
	teamName := acctest.RandomWithPrefix("tf-acc-test-%s")
	description := "tf-acc-group-description-update"
	rn := "github_team_sync_group_mapping.test_mapping"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGithubTeamSyncGroupMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubTeamSyncGroupMappingConfig(teamName),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccGithubTeamSyncGroupMappingImportStateIdFunc(rn),
				ImportStateVerify: true,
			},
			{
				Config: testAccGithubTeamSyncGroupMappingEmptyConfig(teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "group.#", "0"),
				),
			},
			{
				Config: testAccGithubTeamSyncGroupMappingConfig(teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "group.#", "3"),
				),
			},
			{
				Config: testAccGithubTeamSyncGroupMappingAddGroupAndUpdateConfig(teamName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "group.#", "3"),
					resource.TestCheckResourceAttr(rn, "group.1385744695.group_description", description),
					resource.TestCheckResourceAttr(rn, "group.2749525965.group_description", description),
					resource.TestCheckResourceAttr(rn, "group.3830341445.group_description", description),
				),
			},
		},
	})
}

func TestAccGithubTeamSyncGroupMapping_empty(t *testing.T) {
	if isEnterprise != "true" {
		t.Skip("Skipping because `ENTERPRISE_ACCOUNT` is not set or set to false")
	}
	teamName := acctest.RandomWithPrefix("tf-acc-test-%s")
	rn := "github_team_sync_group_mapping.test_mapping"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGithubTeamSyncGroupMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGithubTeamSyncGroupMappingEmptyConfig(teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rn, "group.#", "0"),
				),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccGithubTeamSyncGroupMappingImportStateIdFunc(rn),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGithubTeamSyncGroupMappingDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*Owner).v3client
	orgName := testAccProvider.Meta().(*Owner).name
	ctx := context.TODO()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_team_sync_group_mapping" {
			continue
		}
		slug := rs.Primary.Attributes["team_slug"]
		groupList, resp, err := conn.Teams.ListIDPGroupsForTeamBySlug(ctx, orgName, slug)
		if err == nil {
			if groupList != nil && len(groupList.Groups) > 0 {
				return fmt.Errorf("Team Sync Group Mapping still exists for team slug %s", slug)
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccCheckGithubTeamSyncGroupMappingDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		conn := testAccProvider.Meta().(*Owner).v3client
		orgName := testAccProvider.Meta().(*Owner).name
		slug := rs.Primary.Attributes["team_slug"]

		emptyGroupList := github.IDPGroupList{Groups: []*github.IDPGroup{}}
		_, _, err := conn.Teams.CreateOrUpdateIDPGroupConnectionsBySlug(context.TODO(), orgName, slug, emptyGroupList)

		return err
	}
}

func testAccGithubTeamSyncGroupMappingImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["team_slug"], nil
	}
}

func testAccGithubTeamSyncGroupMappingConfig(teamName string) string {
	return fmt.Sprintf(`
data "github_organization_team_sync_groups" "test_groups" {}

resource "github_team" "test_team" {
  name        = "%s"
  description = "team for acc-tests"
}

resource "github_team_sync_group_mapping" "test_mapping" {
  team_slug  = github_team.test_team.slug

  dynamic "group" {
    for_each = [for g in data.github_organization_team_sync_groups.test_groups.groups : g if length(regexall("^acctest-github-provider", g.group_name)) > 0]
    content {
      group_id          = group.value.group_id
      group_name        = group.value.group_name
      group_description = group.value.group_description
    }
  }
}
`, teamName)
}

func testAccGithubTeamSyncGroupMappingAddGroupAndUpdateConfig(teamName, description string) string {
	return fmt.Sprintf(`
data "github_organization_team_sync_groups" "test_groups" {}

resource "github_team" "test_team" {
  name        = "%s"
  description = "team for acc-tests"
}

resource "github_team_sync_group_mapping" "test_mapping" {
  team_slug  = github_team.test_team.slug

  dynamic "group" {
    for_each = [for g in data.github_organization_team_sync_groups.test_groups.groups : g if length(regexall("^acctest-github-provider", g.group_name)) > 0]
    content {
      group_id          = group.value.group_id
      group_name        = group.value.group_name
      group_description = "%s"
    }
  }
}
`, teamName, description)
}

func testAccGithubTeamSyncGroupMappingEmptyConfig(teamName string) string {
	return fmt.Sprintf(`
data "github_organization_team_sync_groups" "test_groups" {}

resource "github_team" "test_team" {
  name        = "%s"
  description = "team for acc-tests"
}

resource "github_team_sync_group_mapping" "test_mapping" {
  team_slug  = github_team.test_team.slug
}
`, teamName)
}
