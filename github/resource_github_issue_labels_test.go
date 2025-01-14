package github

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubIssueLabels(t *testing.T) {
	t.Run("authoritatively overtakes existing labels", func(t *testing.T) {
		repoName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum))
		existingLabelName := fmt.Sprintf("label-%s", acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum))
		empty := []map[string]interface{}{}

		testCase := func(t *testing.T, mode string) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					// 0. Create the repository
					{
						Config:             testAccGithubIssueLabelsConfig(repoName, nil),
						ExpectNonEmptyPlan: true,
					},
					// 1. Check if some labels already exist (indicated by non-empty plan)
					{
						PreConfig: func() {
							err := testAccGithubIssueLabelsAddLabel(repoName, existingLabelName)
							if err != nil {
								t.Fatalf("failed to add label: %s", existingLabelName)
							}
						},
						Config:             testAccGithubIssueLabelsConfig(repoName, empty),
						PlanOnly:           true,
						ExpectNonEmptyPlan: true,
					},
					// 2. Check if existing labels can be adopted
					{
						Config: testAccGithubIssueLabelsConfig(repoName, append(empty, map[string]interface{}{
							"name":        existingLabelName,
							"color":       "000000",
							"description": "Test label",
						})),
						Check: resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "1"),
					},
					// 3. Check if all the labels are destroyed when the resource has no labels
					{
						Config: testAccGithubIssueLabelsConfig(repoName, empty),
						Check:  resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "0"),
					},
					// 4. Check if a new label can be created
					{
						Config: testAccGithubIssueLabelsConfig(repoName, append(empty, map[string]interface{}{
							"name":        "foo",
							"color":       "000000",
							"description": "foo",
						})),
						Check: resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "1"),
					},
					// 5. Check if a label can be recreated
					{
						Config: testAccGithubIssueLabelsConfig(repoName, append(empty, map[string]interface{}{
							"name":        "Foo",
							"color":       "000000",
							"description": "foo",
						})),
						Check: resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "1"),
					},
					// 6. Check if multiple labels can be created
					{
						Config: testAccGithubIssueLabelsConfig(repoName, append(empty,
							map[string]interface{}{
								"name":        "Foo",
								"color":       "000000",
								"description": "foo",
							},
							map[string]interface{}{
								"name":        "bar",
								"color":       "000000",
								"description": "bar",
							}, map[string]interface{}{
								"name":        "baz",
								"color":       "000000",
								"description": "baz",
							})),
						Check: resource.TestCheckResourceAttr("github_issue_labels.test", "label.#", "3"),
					},
					// 7. Check if labels can be destroyed
					{
						Config:             testAccGithubIssueLabelsConfig(repoName, nil),
						ExpectNonEmptyPlan: true,
					},
					// 8. Check if labels were actually destroyed
					{
						Config:             testAccGithubIssueLabelsConfig(repoName, empty),
						PlanOnly:           true,
						ExpectNonEmptyPlan: true,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			testCase(t, individual)
		})

		t.Run("with an organization account", func(t *testing.T) {
			testCase(t, organization)
		})
	})
}

func testAccGithubIssueLabelsConfig(repoName string, labels []map[string]interface{}) string {
	resource := ""
	if labels != nil {
		dynamic := ""
		for _, label := range labels {
			dynamic += fmt.Sprintf(`
				label {
					name = "%s"
					color = "%s"
					description = "%s"
				}
			`, label["name"], label["color"], label["description"])
		}

		resource = fmt.Sprintf(`
			resource "github_issue_labels" "test" {
				repository = github_repository.test.name

				%s
			}
		`, dynamic)
	}

	return fmt.Sprintf(`
		resource "github_repository" "test" {
			name = "%s"
			auto_init = true
		}

		%s
	`, repoName, resource)
}

func testAccGithubIssueLabelsAddLabel(repository, label string) error {
	client := testAccProvider.Meta().(*Owner).v3client
	orgName := testAccProvider.Meta().(*Owner).name
	ctx := context.TODO()

	_, _, err := client.Issues.CreateLabel(ctx, orgName, repository, &github.Label{Name: github.String(label)})
	return err
}
