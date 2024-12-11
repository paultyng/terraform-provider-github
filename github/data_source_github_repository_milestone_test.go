package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryMilestoneDataSource(t *testing.T) {
	t.Run("queries a repository milestone", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name      = "tf-acc-test-%s"
			}

			resource "github_repository_milestone" "test" {
				owner = split("/", "${github_repository.test.full_name}")[0]
				repository = github_repository.test.name
		    title = "v1.0.0"
		    description = "General Availability"
		    due_date = "2020-11-22"
		    state = "closed"
			}

			data "github_repository_milestone" "test" {
				owner = github_repository_milestone.test.owner
				repository = github_repository_milestone.test.repository
		   	number = github_repository_milestone.test.number
			}

		`, randomID)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_milestone.test", "state",
				"closed",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})
}
