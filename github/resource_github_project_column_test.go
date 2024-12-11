package github

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubProjectColumn_basic(t *testing.T) {
	t.Skip("Skipping test as the GitHub API no longer supports classic projects")

	t.Run("creates and updates a project column", func(t *testing.T) {
		var column github.ProjectColumn

		rn := "github_project_column.column"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccGithubProjectColumnDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubProjectColumnConfig("new column name"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubProjectColumnExists(rn, &column),
						testAccCheckGithubProjectColumnAttributes(&column, &testAccGithubProjectColumnExpectedAttributes{
							Name: "new column name",
						}),
					),
				},
				{
					Config: testAccGithubProjectColumnConfig("updated column name"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubProjectColumnExists(rn, &column),
						testAccCheckGithubProjectColumnAttributes(&column, &testAccGithubProjectColumnExpectedAttributes{
							Name: "updated column name",
						}),
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

func testAccGithubProjectColumnDestroy(s *terraform.State) error {
	meta, err := getTestMeta()
	if err != nil {
		return err
	}
	conn := meta.v3client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_project_column" {
			continue
		}

		columnID, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}

		column, res, err := conn.Projects.GetProjectColumn(context.TODO(), columnID)
		if err == nil {
			if column != nil &&
				column.GetID() == columnID {
				return fmt.Errorf("project column still exists")
			}
		}
		if res.StatusCode != 404 {
			return err
		}
	}
	return nil
}

func testAccCheckGithubProjectColumnExists(n string, project *github.ProjectColumn) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		columnID, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}

		meta, err := getTestMeta()
		if err != nil {
			return err
		}
		conn := meta.v3client

		gotColumn, _, err := conn.Projects.GetProjectColumn(context.TODO(), columnID)
		if err != nil {
			return err
		}
		*project = *gotColumn
		return nil
	}
}

type testAccGithubProjectColumnExpectedAttributes struct {
	Name string
}

func testAccCheckGithubProjectColumnAttributes(column *github.ProjectColumn, want *testAccGithubProjectColumnExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if name := column.GetName(); name != want.Name {
			return fmt.Errorf("got project column %q; want %q", name, want.Name)
		}

		return nil
	}
}

func testAccGithubProjectColumnConfig(columnName string) string {
	return fmt.Sprintf(`
resource "github_organization_project" "test" {
  name = "test-project"
  body = "this is a test project"
}

resource "github_project_column" "column" {
  project_id = "${github_organization_project.test.id}"
  name       = "%s"
}
`, columnName)
}
