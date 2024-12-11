package github

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestSuppressDeployKeyDiff(t *testing.T) {
	testCases := []struct {
		OldValue, NewValue string
		ExpectSuppression  bool
	}{
		{
			"ssh-rsa AAAABB...cd+==",
			"ssh-rsa AAAABB...cd+== terraform-acctest@hashicorp.com\n",
			true,
		},
		{
			"ssh-rsa AAAABB...cd+==",
			"ssh-rsa AAAABB...cd+==",
			true,
		},
		{
			"ssh-rsa AAAABV...cd+==",
			"ssh-rsa DIFFERENT...cd+==",
			false,
		},
	}

	tcCount := len(testCases)
	for i, tc := range testCases {
		suppressed := suppressDeployKeyDiff("test", tc.OldValue, tc.NewValue, nil)
		if tc.ExpectSuppression && !suppressed {
			t.Fatalf("%d/%d: Expected %q and %q to be suppressed",
				i+1, tcCount, tc.OldValue, tc.NewValue)
		}
		if !tc.ExpectSuppression && suppressed {
			t.Fatalf("%d/%d: Expected %q and %q NOT to be suppressed",
				i+1, tcCount, tc.OldValue, tc.NewValue)
		}
	}
}

func TestAccGithubRepositoryDeployKey_basic(t *testing.T) {
	t.Run("creates repository deploy key without error", func(t *testing.T) {
		cmd := exec.Command("bash", "-c", "ssh-keygen -t rsa -b 4096 -C test@example.com -N '' -f test-fixtures/id_rsa>/dev/null <<< y >/dev/null")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		rn := "github_repository_deploy_key.test_repo_deploy_key"
		rs := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		repositoryName := fmt.Sprintf("acctest-%s", rs)
		keyPath := strings.ReplaceAll(filepath.Join("test-fixtures", "id_rsa.pub"), "\\", "/")

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckGithubRepositoryDeployKeyDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccGithubRepositoryDeployKeyConfig(repositoryName, keyPath),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGithubRepositoryDeployKeyExists(rn),
						resource.TestCheckResourceAttr(rn, "read_only", "false"),
						resource.TestCheckResourceAttr(rn, "repository", repositoryName),
						resource.TestMatchResourceAttr(rn, "key", regexp.MustCompile(`^ssh-rsa [^\s]+$`)),
						resource.TestCheckResourceAttr(rn, "title", "title"),
					),
				},
			},
		})
	})
}

func testAccCheckGithubRepositoryDeployKeyDestroy(s *terraform.State) error {
	meta, err := getTestMeta()
	if err != nil {
		return err
	}
	conn := meta.v3client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "github_repository_deploy_key" {
			continue
		}

		owner := meta.name
		repoName, idString, err := parseTwoPartID(rs.Primary.ID, "repository", "ID")
		if err != nil {
			return err
		}

		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			return unconvertibleIdErr(idString, err)
		}

		_, resp, err := conn.Repositories.GetKey(context.TODO(), owner, repoName, id)

		if err != nil && resp.Response.StatusCode != 404 {
			return err
		}
		return nil
	}

	return nil
}

func testAccCheckGithubRepositoryDeployKeyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no membership ID is set")
		}

		meta, err := getTestMeta()
		if err != nil {
			return err
		}
		conn := meta.v3client
		owner := meta.name
		repoName, idString, err := parseTwoPartID(rs.Primary.ID, "repository", "ID")
		if err != nil {
			return err
		}

		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			return unconvertibleIdErr(idString, err)
		}

		_, _, err = conn.Repositories.GetKey(context.TODO(), owner, repoName, id)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccGithubRepositoryDeployKeyConfig(name, keyPath string) string {
	return fmt.Sprintf(`
resource "github_repository" "test_repo" {
  name = "%s"
}

resource "github_repository_deploy_key" "test_repo_deploy_key" {
  key        = "${file("%s")}"
  read_only  = "false"
  repository = "${github_repository.test_repo.name}"
  title      = "title"
}
`, name, keyPath)
}
