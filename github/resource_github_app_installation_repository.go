package github

import (
	"context"
	"log"
	"strconv"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubAppInstallationRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubAppInstallationRepositoryCreate,
		Read:   resourceGithubAppInstallationRepositoryRead,
		Delete: resourceGithubAppInstallationRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"installation_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repo_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceGithubAppInstallationRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	installationIDString := d.Get("installation_id").(string)
	installationID, err := strconv.ParseInt(installationIDString, 10, 64)
	if err != nil {
		return unconvertibleIdErr(installationIDString, err)
	}

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()
	repoName := d.Get("repository").(string)
	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return err
	}
	repoID := repo.GetID()

	log.Printf("[DEBUG] Creating app installation repository association: %s:%s",
		installationIDString, repoName)

	_, _, err = client.Apps.AddRepository(ctx, installationID, repoID)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(installationIDString, repoName))
	return resourceGithubAppInstallationRepositoryRead(d, meta)
}

func resourceGithubAppInstallationRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	installationIDString, repoName, err := parseTwoPartID(d.Id(), "installation_id", "repository")
	if err != nil {
		return err
	}

	installationID, err := strconv.ParseInt(installationIDString, 10, 64)
	if err != nil {
		return unconvertibleIdErr(installationIDString, err)
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	opt := &github.ListOptions{PerPage: maxPerPage}

	for {
		repos, resp, err := client.Apps.ListUserRepos(ctx, installationID, opt)
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Found %d repos, checking if any matches %s", repos.TotalCount, repoName)
		for _, r := range repos.Repositories {
			if r.GetName() == repoName {
				d.Set("installation_id", installationIDString)
				d.Set("repository", repoName)
				d.Set("repo_id", r.GetID())
				return nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	log.Printf("[WARN] Removing app installation repository association %s from state because it no longer exists in GitHub",
		d.Id())
	d.SetId("")
	return nil
}

func resourceGithubAppInstallationRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}
	installationIDString := d.Get("installation_id").(string)
	installationID, err := strconv.ParseInt(installationIDString, 10, 64)
	if err != nil {
		return unconvertibleIdErr(installationIDString, err)
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()

	repoName := d.Get("repository").(string)
	repoID := d.Get("repo_id").(int)
	log.Printf("[DEBUG] Deleting app installation repository association: %s:%s",
		installationIDString, repoName)

	_, err = client.Apps.RemoveRepository(ctx, installationID, int64(repoID))
	if err != nil {
		return err
	}
	return nil
}
