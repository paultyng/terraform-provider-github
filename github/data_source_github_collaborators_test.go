package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubCollaboratorsDataSource(t *testing.T) {
	t.Run("gets all collaborators", func(t *testing.T) {
		repoName := fmt.Sprintf("tf-acc-test-collab-%s", acctest.RandString(5))
		config := fmt.Sprintf(`
resource "github_repository" "test" {
  name = "%s"
}

data "github_collaborators" "test" {
  owner      = "%s"
  repository = "${github_repository.test.name}"
}
`, repoName, testAccConf.owner)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_collaborators.test", "collaborator.#"),
			resource.TestCheckResourceAttr("data.github_collaborators.test", "affiliation", "all"),
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

	t.Run("gets admin collaborators", func(t *testing.T) {
		repoName := fmt.Sprintf("tf-acc-test-collab-%s", acctest.RandString(5))
		config := fmt.Sprintf(`
resource "github_repository" "test" {
  name = "%s"
}

data "github_collaborators" "test" {
  owner      = "%s"
  repository = "${github_repository.test.name}"
  permission = "admin"
}
`, repoName, testAccConf.owner)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_collaborators.test", "collaborator.#"),
			resource.TestCheckResourceAttr("data.github_collaborators.test", "affiliation", "all"),
			resource.TestCheckResourceAttr("data.github_collaborators.test", "permission", "admin"),
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
