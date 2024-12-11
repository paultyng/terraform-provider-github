package github

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGithubRepositoryCollaborators(t *testing.T) {
	if len(testAccConf.testExternalUser) == 0 {
		t.Skip("No external user provided")
	}

	meta, err := getTestMeta()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("adds user collaborator", func(t *testing.T) {
		conn := meta.v3client
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)

		config := fmt.Sprintf(`
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
		`, repoName, testAccConf.testExternalUser)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "user.#"),
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "user.#", "1"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "0"),
						func(state *terraform.State) error {
							owner := testAccConf.owner

							collaborators := state.RootModule().Resources["github_repository_collaborators.test_repo_collaborators"].Primary
							for name, val := range collaborators.Attributes {
								if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".username") && val != testAccConf.testExternalUser {
									return fmt.Errorf("expected user.*.username to be set to %s, was %s", testAccConf.testExternalUser, val)
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
								return fmt.Errorf("expected an invite for %s but not found", testAccConf.testExternalUser)
							}
							if invites[0].GetInvitee().GetLogin() != testAccConf.testExternalUser {
								return fmt.Errorf("expected an invite for %s for repo %s/%s", testAccConf.testExternalUser, owner, repoName)
							}
							perm := invites[0].GetPermissions()
							if perm != "write" {
								return fmt.Errorf("expected the invite for %s to have push perms for for %s/%s, found %s", testAccConf.testExternalUser, owner, repoName, perm)
							}
							return nil
						},
					),
				},
			},
		})
	})

	t.Run("adds team collaborator", func(t *testing.T) {
		conn := meta.v3client
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test" {
				name = "test"
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
		`, repoName, testAccConf.testExternalUser)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "user.#"),
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "user.#", "1"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "1"),
						func(state *terraform.State) error {
							owner := testAccConf.owner

							teamAttrs := state.RootModule().Resources["github_team.test"].Primary.Attributes
							collaborators := state.RootModule().Resources["github_repository_collaborators.test_repo_collaborators"].Primary
							for name, val := range collaborators.Attributes {
								if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".username") && val != testAccConf.testExternalUser {
									return fmt.Errorf("expected user.*.username to be set to %s, was %s", testAccConf.testExternalUser, val)
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
								return fmt.Errorf("expected %s to be a collaborator for repo %s/%s", testAccConf.testExternalUser, owner, repoName)
							}
							perm := getPermission(users[0].GetRoleName())
							if perm != "admin" {
								return fmt.Errorf("expected %s to have admin perms for repo %s/%s, found %s", testAccConf.testExternalUser, owner, repoName, perm)
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
					),
				},
			},
		})
	})

	t.Run("updates user collaborators without error", func(t *testing.T) {
		if len(testAccConf.testExternalUser2) == 0 {
			t.Skip("No additional external user provided")
		}

		conn := meta.v3client
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)

		config := fmt.Sprintf(`
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
		`, repoName, testAccConf.testExternalUser)

		configUpdate := fmt.Sprintf(`
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
		`, repoName, testAccConf.testExternalUser2)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					Config: configUpdate,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "user.#"),
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "user.#", "1"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "0"),
						func(state *terraform.State) error {
							owner := testAccConf.owner

							collaborators := state.RootModule().Resources["github_repository_collaborators.test_repo_collaborators"].Primary
							for name, val := range collaborators.Attributes {
								if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".username") && val != testAccConf.testExternalUser2 {
									return fmt.Errorf("expected user.*.username to be set to %s, was %s", testAccConf.testExternalUser, val)
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
								return fmt.Errorf("expected an invite for %s but not found", testAccConf.testExternalUser)
							}
							if invites[0].GetInvitee().GetLogin() != testAccConf.testExternalUser2 {
								return fmt.Errorf("expected an invite for %s for repo %s/%s", testAccConf.testExternalUser, owner, repoName)
							}
							perm := getPermission(invites[0].GetPermissions())
							if perm != "pull" {
								return fmt.Errorf("expected the invite for %s to have pull perms for for %s/%s, found %s", testAccConf.testExternalUser, owner, repoName, perm)
							}
							return nil
						},
					),
				},
			},
		})
	})

	t.Run("updates team collaborators without error", func(t *testing.T) {
		if len(testAccConf.testExternalUser2) == 0 {
			t.Skip("No additional external user provided")
		}

		conn := meta.v3client
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test" {
				name = "test"
			}

			resource "github_team" "test2" {
				name = "test2"
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
		`, repoName, testAccConf.testExternalUser, testAccConf.testExternalUser2)

		configUpdate := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test" {
				name = "test"
			}

			resource "github_team" "test2" {
				name = "test2"
			}

			resource "github_repository_collaborators" "test_repo_collaborators" {
				repository = "${github_repository.test.name}"

				user {
					username   = "%s"
					permission = "push"
				}
				team {
					team_id   = github_team.test.id
					permission = "push"
				}
			}
		`, repoName, testAccConf.testExternalUser)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					Config: configUpdate,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "user.#"),
						resource.TestCheckResourceAttrSet("github_repository_collaborators.test_repo_collaborators", "team.#"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "user.#", "1"),
						resource.TestCheckResourceAttr("github_repository_collaborators.test_repo_collaborators", "team.#", "1"),
						func(state *terraform.State) error {
							owner := testAccConf.owner

							teamAttrs := state.RootModule().Resources["github_team.test"].Primary.Attributes
							collaborators := state.RootModule().Resources["github_repository_collaborators.test_repo_collaborators"].Primary
							for name, val := range collaborators.Attributes {
								if strings.HasPrefix(name, "user.") && strings.HasSuffix(name, ".username") && val != testAccConf.testExternalUser {
									return fmt.Errorf("expected user.*.username to be set to %s, was %s", testAccConf.testExternalUser, val)
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
								return fmt.Errorf("expected %s to be a collaborator for repo %s/%s", testAccConf.testExternalUser, owner, repoName)
							}
							perm := getPermission(users[0].GetRoleName())
							if perm != "push" {
								return fmt.Errorf("expected %s to have push perms for repo %s/%s, found %s", testAccConf.testExternalUser, owner, repoName, perm)
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
					),
				},
			},
		})
	})

	t.Run("removes user collaborators without error", func(t *testing.T) {
		conn := meta.v3client
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)

		config := fmt.Sprintf(`
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
		`, repoName, testAccConf.testExternalUser)

		configUpdate := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					Config: configUpdate,
					Check: resource.ComposeTestCheckFunc(
						func(state *terraform.State) error {
							owner := testAccConf.owner

							invites, _, err := conn.Repositories.ListInvitations(context.TODO(), owner, repoName, nil)
							if err != nil {
								return err
							}
							if len(invites) != 0 {
								return fmt.Errorf("expected no invites but not found %d", len(invites))
							}
							return nil
						},
					),
				},
			},
		})
	})

	t.Run("removes team collaborators without error", func(t *testing.T) {
		if len(testAccConf.testExternalUser2) == 0 {
			t.Skip("No additional external user provided")
		}

		conn := meta.v3client
		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("tf-acc-test-%s", randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test" {
				name = "test"
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
		`, repoName, testAccConf.testExternalUser, testAccConf.testExternalUser2)

		configUpdate := fmt.Sprintf(`
			resource "github_repository" "test" {
				name = "%s"
				auto_init = true
				visibility = "private"
			}

			resource "github_team" "test" {
				name = "test"
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnlessHasOrgs(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					Config: configUpdate,
					Check: resource.ComposeTestCheckFunc(
						func(state *terraform.State) error {
							owner := testAccConf.owner

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
					),
				},
			},
		})
	})
}
