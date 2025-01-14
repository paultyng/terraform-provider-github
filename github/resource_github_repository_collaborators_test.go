package github

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubRepositoryCollaborators(t *testing.T) {
	inOrgUser := os.Getenv("GITHUB_IN_ORG_USER")
	inOrgUser2 := os.Getenv("GITHUB_IN_ORG_USER2")

	config := Config{BaseURL: "https://api.github.com/", Owner: testOwnerFunc(), Token: testToken}
	meta, err := config.Meta()
	if err != nil {
		t.Fatalf("failed to return meta without error: %s", err.Error())
	}

	t.Run("creates collaborators without error", func(t *testing.T) {
		if inOrgUser == "" {
			t.Skip("set inOrgUser to unskip this test run")
		}

		if inOrgUser == testOwnerFunc() {
			t.Skip("inOrgUser can't be same as owner")
		}

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		conn := meta.(*Owner).v3client
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)
		teamName := fmt.Sprintf("tf-acc-test-team-%s", randomID)

		individualConfig := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "push"
				}
			}
		`, repoName, inOrgUser)

		orgConfig := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test" {
				name = "%s"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "admin"
				}
				team {
					team_id   = github_team.test.id
					permission = "pull"
				}
			}
		`, repoName, teamName, inOrgUser)

		testCase := func(t *testing.T, mode, config string, testCheck func(state *terraform.State) error) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  testCheck,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			check := resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "user.#"),
				resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
				resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "user.#", "1"),
				resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "0"),
				func(state *terraform.State) error {
					owner := meta.(*Owner).name

					collaborators := state.RootModule().Resources["github_repository_collaborators.test_repo_collaborators"].Primary
					for name, val := range collaborators.Attributes {
						if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".username") && val != inOrgUser {
							return fmt.Errorf("expected user.*.username to be set to %s, was %s", inOrgUser, val)
						}
						if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".permission") && val != "push" {
							return fmt.Errorf("expected user.*.permission to be set to push, was %s", val)
						}
					}

					invites, _, err := conn.Repositories.ListInvitations(context.TODO(), owner, repoName, nil)
					if err != nil {
						return err
					}
					if len(invites) != 1 {
						return fmt.Errorf("expected an invite for %s but not found", inOrgUser)
					}
					if invites[0].GetInvitee().GetLogin() != inOrgUser {
						return fmt.Errorf("expected an invite for %s for repo %s/%s", inOrgUser, owner, repoName)
					}
					perm := invites[0].GetPermissions()
					if perm != "write" {
						return fmt.Errorf("expected the invite for %s to have push perms for for %s/%s, found %s", inOrgUser, owner, repoName, perm)
					}
					return nil
				},
			)
			testCase(t, individual, individualConfig, check)
		})

		t.Run("with an organization account", func(t *testing.T) {
			check := resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "user.#"),
				resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
				resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "user.#", "1"),
				resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "1"),
				func(state *terraform.State) error {
					owner := testOrganizationFunc()

					teamAttrs := state.RootModule().Resources["github_team.test"].Primary.Attributes
					collaborators := state.RootModule().Resources["github_repository_collaborators.test_repo_collaborators"].Primary
					for name, val := range collaborators.Attributes {
						if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".username") && val != inOrgUser {
							return fmt.Errorf("expected user.*.username to be set to %s, was %s", inOrgUser, val)
						}
						if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".permission") && val != "admin" {
							return fmt.Errorf("expected user.*.permission to be set to admin, was %s", val)
						}
						if strings.HasPrefix(name, "team.") && strings.HasSuffix(name, ".team_id") && val != teamAttrs["id"] {
							return fmt.Errorf("expected team.*.team_id to be set to %s, was %s", teamAttrs["id"], val)
						}
						if strings.HasPrefix(name, "team.") && strings.HasSuffix(name, ".permission") && val != "pull" {
							return fmt.Errorf("expected team.*.permission to be set to pull, was %s", val)
						}
					}
					users, _, err := conn.Repositories.ListCollaborators(context.TODO(), owner, repoName, &github.ListCollaboratorsOptions{Affiliation: "direct"})
					if err != nil {
						return err
					}
					if len(users) != 1 {
						return fmt.Errorf("expected %s to be a collaborator for repo %s/%s", inOrgUser, owner, repoName)
					}
					perm := getPermission(users[0].GetRoleName())
					if perm != "admin" {
						return fmt.Errorf("expected %s to have admin perms for repo %s/%s, found %s", inOrgUser, owner, repoName, perm)
					}
					teams, _, err := conn.Repositories.ListTeams(context.TODO(), owner, repoName, nil)
					if err != nil {
						return err
					}
					if len(teams) != 1 {
						return fmt.Errorf("expected team %s to be a collaborator for %s/%s", repoName, owner, repoName)
					}
					perm = getPermission(teams[0].GetPermission())
					if perm != "pull" {
						return fmt.Errorf("expected team %s to have pull perms for repo %s/%s, found %s", repoName, owner, repoName, perm)
					}
					return nil
				},
			)
			testCase(t, organization, orgConfig, check)
		})
	})

	t.Run("updates collaborators without error", func(t *testing.T) {
		if inOrgUser == "" || inOrgUser2 == "" {
			t.Skip("set inOrgUser and inOrgUser2 to unskip this test run")
		}

		if inOrgUser == testOwnerFunc() || inOrgUser2 == testOwnerFunc() {
			t.Skip("inOrgUser or inOrgUser2 can't be same as owner")
		}

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		conn := meta.(*Owner).v3client
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)
		team0Name := fmt.Sprintf("tf-acc-test-team-0-%s", randomID)
		team1Name := fmt.Sprintf("tf-acc-test-team-1-%s", randomID)

		individualConfig := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "push"
				}
			}
		`, repoName, inOrgUser)

		individualConfigUpdate := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "pull"
				}
			}
		`, repoName, inOrgUser2)

		orgConfig := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test_0" {
				name = "%s"
			}

			resource "github_team" "test_1" {
				name = "%s"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "admin"
				}
				user {
					username   = "%s"
					permission = "admin"
				}
				team {
					team_id   = github_team.test.id
					permission = "pull"
				}
				team {
					team_id   = github_team.test2.id
					permission = "pull"
				}
			}
		`, repoName, team0Name, team1Name, inOrgUser, inOrgUser2)

		orgConfigUpdate := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test_0" {
				name = "%s"
			}

			resource "github_team" "test_1" {
				name = "%s"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "push"
				}
				team {
					team_id   = github_team.test_0.id
					permission = "push"
				}
			}
		`, repoName, team0Name, team1Name, inOrgUser)

		testCase := func(t *testing.T, mode, config, configUpdate string, testCheck func(state *terraform.State) error) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
					},
					{
						Config: configUpdate,
						Check:  testCheck,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			check := resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "user.#"),
				resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "user.#", "1"),
				func(state *terraform.State) error {
					owner := meta.(*Owner).name

					collaborators := state.RootModule().Resources["github_repository_collaborators.test_repo_collaborators"].Primary
					for name, val := range collaborators.Attributes {
						if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".username") && val != inOrgUser2 {
							return fmt.Errorf("expected user.*.username to be set to %s, was %s", inOrgUser, val)
						}
						if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".permission") && val != "pull" {
							return fmt.Errorf("expected user.*.permission to be set to pull, was %s", val)
						}
					}

					invites, _, err := conn.Repositories.ListInvitations(context.TODO(), owner, repoName, nil)
					if err != nil {
						return err
					}
					if len(invites) != 1 {
						return fmt.Errorf("expected an invite for %s but not found", inOrgUser)
					}
					if invites[0].GetInvitee().GetLogin() != inOrgUser2 {
						return fmt.Errorf("expected an invite for %s for repo %s/%s", inOrgUser, owner, repoName)
					}
					perm := getPermission(invites[0].GetPermissions())
					if perm != "pull" {
						return fmt.Errorf("expected the invite for %s to have pull perms for for %s/%s, found %s", inOrgUser, owner, repoName, perm)
					}
					return nil
				},
			)
			testCase(t, individual, individualConfig, individualConfigUpdate, check)
		})

		t.Run("with an organization account", func(t *testing.T) {
			check := resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "user.#"),
				resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
				resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "user.#", "1"),
				resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "1"),
				func(state *terraform.State) error {
					owner := testOrganizationFunc()

					teamAttrs := state.RootModule().Resources["github_team.test_0"].Primary.Attributes
					collaborators := state.RootModule().Resources["github_repository_collaborators.test_repo_collaborators"].Primary
					for name, val := range collaborators.Attributes {
						if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".username") && val != inOrgUser {
							return fmt.Errorf("expected user.*.username to be set to %s, was %s", inOrgUser, val)
						}
						if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".permission") && val != "push" {
							return fmt.Errorf("expected user.*.permission to be set to push, was %s", val)
						}
						if strings.HasPrefix(name, "team.") && strings.HasSuffix(name, ".team_id") && val != teamAttrs["id"] {
							return fmt.Errorf("expected team.*.team_id to be set to %s, was %s", teamAttrs["id"], val)
						}
						if strings.HasPrefix(name, "team.") && strings.HasSuffix(name, ".permission") && val != "push" {
							return fmt.Errorf("expected team.*.permission to be set to push, was %s", val)
						}
					}

					users, _, err := conn.Repositories.ListCollaborators(context.TODO(), owner, repoName, &github.ListCollaboratorsOptions{Affiliation: "direct"})
					if err != nil {
						return err
					}
					if len(users) != 1 {
						return fmt.Errorf("expected %s to be a collaborator for repo %s/%s", inOrgUser, owner, repoName)
					}
					perm := getPermission(users[0].GetRoleName())
					if perm != "push" {
						return fmt.Errorf("expected %s to have push perms for repo %s/%s, found %s", inOrgUser, owner, repoName, perm)
					}
					teams, _, err := conn.Repositories.ListTeams(context.TODO(), owner, repoName, nil)
					if err != nil {
						return err
					}
					if len(teams) != 1 {
						return fmt.Errorf("expected team %s to be a collaborator for %s/%s", repoName, owner, repoName)
					}
					perm = getPermission(teams[0].GetPermission())
					if perm != "push" {
						return fmt.Errorf("expected team %s to have push perms for repo %s/%s, found %s", repoName, owner, repoName, perm)
					}
					return nil
				},
			)
			testCase(t, organization, orgConfig, orgConfigUpdate, check)
		})
	})

	t.Run("removes collaborators without error", func(t *testing.T) {
		if inOrgUser == "" || inOrgUser2 == "" {
			t.Skip("set inOrgUser and inOrgUser2 to unskip this test run")
		}

		if inOrgUser == testOwnerFunc() || inOrgUser2 == testOwnerFunc() {
			t.Skip("inOrgUser or inOrgUser2 can't be same as owner")
		}

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		conn := meta.(*Owner).v3client
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)
		teamName := fmt.Sprintf("tf-acc-test-team-%s", randomID)

		individualConfig := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "push"
				}
			}
		`, repoName, inOrgUser)

		individualConfigUpdate := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}
		`, repoName)

		orgConfig := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test" {
				name = "%s"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "admin"
				}
				user {
					username   = "%s"
					permission = "admin"
				}
				team {
					team_id   = github_team.test.id
					permission = "pull"
				}
			}
		`, repoName, teamName, inOrgUser, inOrgUser2)

		orgConfigUpdate := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test" {
				name = "%s"
			}
		`, repoName, teamName)

		testCase := func(t *testing.T, mode, config, configUpdate string, testCheck func(state *terraform.State) error) {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { skipUnlessMode(t, mode) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: config,
					},
					{
						Config: configUpdate,
						Check:  testCheck,
					},
				},
			})
		}

		t.Run("with an anonymous account", func(t *testing.T) {
			t.Skip("anonymous account not supported for this operation")
		})

		t.Run("with an individual account", func(t *testing.T) {
			check := resource.ComposeTestCheckFunc(
				func(state *terraform.State) error {
					owner := meta.(*Owner).name

					invites, _, err := conn.Repositories.ListInvitations(context.TODO(), owner, repoName, nil)
					if err != nil {
						return err
					}
					if len(invites) != 0 {
						return fmt.Errorf("expected no invites but not found %d", len(invites))
					}
					return nil
				},
			)
			testCase(t, individual, individualConfig, individualConfigUpdate, check)
		})

		t.Run("with an organization account", func(t *testing.T) {
			check := resource.ComposeTestCheckFunc(
				func(state *terraform.State) error {
					owner := testOrganizationFunc()

					users, _, err := conn.Repositories.ListCollaborators(context.TODO(), owner, repoName, &github.ListCollaboratorsOptions{Affiliation: "direct"})
					if err != nil {
						return err
					}
					if len(users) != 0 {
						return fmt.Errorf("expected no collaborators for repo %s/%s but found %d", owner, repoName, len(users))
					}
					teams, _, err := conn.Repositories.ListTeams(context.TODO(), owner, repoName, nil)
					if err != nil {
						return err
					}
					if len(teams) != 0 {
						return fmt.Errorf("expected no teams to be a collaborator for %s/%s but found %d", owner, repoName, len(teams))
					}
					return nil
				},
			)
			testCase(t, organization, orgConfig, orgConfigUpdate, check)
		})
	})

	t.Run("does not churn on team slug", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)
		team0Name := fmt.Sprintf("tf-acc-test-team-0-%s", randomID)
		team1Name := fmt.Sprintf("tf-acc-test-team-1-%s", randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test_0" {
				name = "%s"
			}

			resource "github_team" "test_1" {
				name = "%s"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				team {
					team_id   = github_team.test_0.id
					permission = "pull"
				}

				team {
					team_id = github_team.test_1.name
					permission = "pull"
				}
			}
		`, repoName, team0Name, team1Name)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "2"),
					),
				},
				{
					Config:             config,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ignores specified teams", func(t *testing.T) {
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)
		team0Name := fmt.Sprintf("tf-acc-test-team-0-%s", randomID)
		team1Name := fmt.Sprintf("tf-acc-test-team-1-%s", randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test_0" {
				name = "%s"
			}

			resource "github_team_repository" "some_team_repo" {
				team_id    = github_team.test_0.id
				repository = github_repository.test.name
			}

			resource "github_team" "test_1" {
				name = "%s"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				team {
					team_id   = github_team.test_1.id
					permission = "pull"
				}

				ignore_team {
					team_id = github_team.test_0.id
				}
			}
		`, repoName, team0Name, team1Name)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { skipUnlessMode(t, organization) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "1"),
					),
				},
				{
					Config:             config,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}
