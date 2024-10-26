package github

import (
	"context"
	"log"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubAppInstallationRepositories() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubAppInstallationRepositoriesCreateOrUpdate,
		Read:   resourceGithubAppInstallationRepositoriesRead,
		Update: resourceGithubAppInstallationRepositoriesCreateOrUpdate,
		Delete: resourceGithubAppInstallationRepositoriesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"installation_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub app installation id.",
			},
			"selected_repositories": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:         schema.HashString,
				Required:    true,
				Description: "A list of repository names to install the app on.",
			},
		},
	}
}

func resourceGithubAppInstallationRepositoriesCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	installationIDString := d.Get("installation_id").(string)
	selectedRepositories := d.Get("selected_repositories")

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, installationIDString)

	selectedRepositoryNames := []string{}

	names := selectedRepositories.(*schema.Set).List()
	for _, name := range names {
		selectedRepositoryNames = append(selectedRepositoryNames, name.(string))
	}

	currentReposNameIDs, instID, err := getAllAccessibleRepos(meta, installationIDString)
	if err != nil {
		return err
	}

	// Add repos that are not in the current state on GitHub
	for _, repoName := range selectedRepositoryNames {
		if _, ok := currentReposNameIDs[repoName]; ok {
			// If it already exists, remove it from the map so we can delete all that are left at the end
			delete(currentReposNameIDs, repoName)
		} else {
			repo, _, err := client.Repositories.Get(ctx, owner, repoName)
			if err != nil {
				return err
			}
			repoID := repo.GetID()
			log.Printf("[DEBUG]: Adding %v:%v to app installation %v", repoName, repoID, instID)
			_, _, err = client.Apps.AddRepository(ctx, instID, repoID)
			if err != nil {
				return err
			}
		}
	}

	// Remove repositories that existed on GitHub but not selectedRepositories
	// There is a github limitation that means we can't remove the last repository from an installation.
	// Therefore, we skip the first and delete the rest. The app will then need to be uninstalled via the GUI
	// as there is no current API endpoint for [un]installation. Ensure there is at least one repository remaining.
	if len(selectedRepositoryNames) >= 1 {
		for repoName, repoID := range currentReposNameIDs {
			log.Printf("[DEBUG]: Removing %v:%v from app installation %v", repoName, repoID, instID)
			_, err = client.Apps.RemoveRepository(ctx, instID, repoID)
			if err != nil {
				return err
			}
		}
	}

	d.SetId(installationIDString)
	return resourceGithubAppInstallationRepositoriesRead(d, meta)
}

func resourceGithubAppInstallationRepositoriesRead(d *schema.ResourceData, meta interface{}) error {
	installationIDString := d.Id()

	reposNameIDs, _, err := getAllAccessibleRepos(meta, installationIDString)
	if err != nil {
		return err
	}

	repoNames := []string{}
	for name := range reposNameIDs {
		repoNames = append(repoNames, name)
	}

	if len(reposNameIDs) > 0 {
		if err = d.Set("installation_id", installationIDString); err != nil {
			return err
		}
		if err = d.Set("selected_repositories", repoNames); err != nil {
			return err
		}
		return nil
	}

	log.Printf("[INFO] Removing app installation repository association %s from state because it no longer exists in GitHub",
		d.Id())
	d.SetId("")
	return nil
}

func resourceGithubAppInstallationRepositoriesDelete(d *schema.ResourceData, meta interface{}) error {
	installationIDString := d.Get("installation_id").(string)

	reposNameIDs, instID, err := getAllAccessibleRepos(meta, installationIDString)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, installationIDString)

	// There is a github limitation that means we can't remove the last repository from an installation.
	// Therefore, we skip the first and delete the rest. The app will then need to be uninstalled via the GUI
	// as there is no current API endpoint for [un]installation.
	first := true
	for repoName, repoID := range reposNameIDs {
		if first {
			first = false
			log.Printf("[WARN]: Cannot remove %v:%v from app installation %v as there must remain at least one repository selected due to API limitations. Manually uninstall the app to remove.", repoName, repoID, instID)
			continue
		} else {
			_, err = client.Apps.RemoveRepository(ctx, instID, repoID)
			log.Printf("[DEBUG]: Removing %v:%v from app installation %v", repoName, repoID, instID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getAllAccessibleRepos(meta interface{}, idString string) (map[string]int64, int64, error) {
	err := checkOrganization(meta)
	if err != nil {
		return nil, 0, err
	}

	installationID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return nil, 0, unconvertibleIdErr(idString, err)
	}

	ctx := context.WithValue(context.Background(), ctxId, idString)
	opt := &github.ListOptions{PerPage: maxPerPage}
	client := meta.(*Owner).v3client

	allRepos := make(map[string]int64)

	for {
		repos, resp, err := client.Apps.ListUserRepos(ctx, installationID, opt)
		if err != nil {
			return nil, 0, err
		}
		for _, r := range repos.Repositories {
			allRepos[r.GetName()] = r.GetID()
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, installationID, nil
}
