package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGithubRepositoryCollaborator(t *testing.T) {
	if len(testAccConf.testExternalUser) == 0 {
		t.Skip("No external user provided")
	}

	t.Run("creates invitations without error", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_repository_collaborator" "test_repo_collaborator" {
				repository = "${github_repository.test.name}"
				username   = "%s"
				permission = "triage"
			}
		`, randomID, testAccConf.testExternalUser)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_collaborator.test_repo_collaborator", "permission",
				"triage",
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

	t.Run("creates invitations when repository contains the org name", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		configWithOwner := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "tf-acc-test-%s"
				auto_init = true
			}

			resource "github_repository_collaborator" "test_repo_collaborator_2" {
				repository = "%s/${github_repository.test.name}"
				username   = "<username>"
				permission = "triage"
			}
		`, randomID, testAccConf.owner)

		checkWithOwner := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(
				"github_repository_collaborator.test_repo_collaborator_2", "permission",
				"triage",
			),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: configWithOwner,
					Check:  checkWithOwner,
				},
			},
		})
	})
}

func TestParseRepoName(t *testing.T) {
	tests := []struct {
		name         string
		repoName     string
		defaultOwner string
		wantOwner    string
		wantRepoName string
	}{
		{
			name:         "Repo name without owner",
			repoName:     "example-repo",
			defaultOwner: "default-owner",
			wantOwner:    "default-owner",
			wantRepoName: "example-repo",
		},
		{
			name:         "Repo name with owner",
			repoName:     "owner-name/example-repo",
			defaultOwner: "default-owner",
			wantOwner:    "owner-name",
			wantRepoName: "example-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepoName := parseRepoName(tt.repoName, tt.defaultOwner)
			if gotOwner != tt.wantOwner || gotRepoName != tt.wantRepoName {
				t.Errorf("parseRepoName(%q, %q) = %q, %q, want %q, %q",
					tt.repoName, tt.defaultOwner, gotOwner, gotRepoName, tt.wantOwner, tt.wantRepoName)
			}
		})
	}
}
