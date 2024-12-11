package github

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryDeployKeysDataSource(t *testing.T) {
	t.Run("manages deploy keys", func(t *testing.T) {
		keyPath := strings.ReplaceAll(filepath.Join("test-fixtures", "id_rsa.pub"), "\\", "/")

		repoName := fmt.Sprintf("tf-acc-test-deploy-keys-%s", acctest.RandString(5))
		config := fmt.Sprintf(`
		resource "github_repository" "test" {
			name      = "%s"
			auto_init = true
		}

		resource "github_repository_deploy_key" "test" {
			repository = github_repository.test.name
			key        = file("%s")
			title      = "title1"
			read_only  = true
		}

		data "github_repository_deploy_keys" "test" {
			repository = github_repository.test.name

			depends_on = [github_repository_deploy_key.test]
		}
		`, repoName, keyPath)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.github_repository_deploy_keys.test", "keys.#"),
						resource.TestCheckResourceAttr("data.github_repository_deploy_keys.test", "keys.#", "1"),
						resource.TestCheckResourceAttr("data.github_repository_deploy_keys.test", "keys.0.title", "title1"),
						resource.TestCheckResourceAttrSet("data.github_repository_deploy_keys.test", "keys.0.id"),
						resource.TestCheckResourceAttr("data.github_repository_deploy_keys.test", "keys.0.verified", "true"),
					),
				},
			},
		})
	})
}
