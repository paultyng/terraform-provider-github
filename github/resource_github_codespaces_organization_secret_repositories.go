package github

import (
	"context"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubCodespacesOrganizationSecretRepositories() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubCodespaceOrganizationSecretRepositoriesCreateOrUpdate,
		Read:   resourceGithubCodespaceOrganizationSecretRepositoriesRead,
		Update: resourceGithubCodespaceOrganizationSecretRepositoriesCreateOrUpdate,
		Delete: resourceGithubCodespaceOrganizationSecretRepositoriesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"secret_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Name of the existing secret.",
				ValidateDiagFunc: validateSecretNameFunc,
			},
			"selected_repository_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Set:         schema.HashInt,
				Required:    true,
				Description: "An array of repository ids that can access the organization secret.",
			},
		},
	}
}

func resourceGithubCodespaceOrganizationSecretRepositoriesCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	secretName := d.Get("secret_name").(string)
	selectedRepositories := d.Get("selected_repository_ids")

	selectedRepositoryIDs := []int64{}

	ids := selectedRepositories.(*schema.Set).List()
	for _, id := range ids {
		selectedRepositoryIDs = append(selectedRepositoryIDs, int64(id.(int)))
	}

	_, err = client.Codespaces.SetSelectedReposForOrgSecret(ctx, owner, secretName, selectedRepositoryIDs)
	if err != nil {
		return err
	}

	d.SetId(secretName)
	return resourceGithubCodespaceOrganizationSecretRepositoriesRead(d, meta)
}

func resourceGithubCodespaceOrganizationSecretRepositoriesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	selectedRepositoryIDs := github.SelectedRepoIDs{}
	opt := &github.ListOptions{
		PerPage: maxPerPage,
	}
	for {
		results, resp, err := client.Codespaces.ListSelectedReposForOrgSecret(ctx, owner, d.Id(), opt)
		if err != nil {
			return err
		}

		for _, repo := range results.Repositories {
			selectedRepositoryIDs = append(selectedRepositoryIDs, repo.GetID())
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	d.Set("selected_repository_ids", selectedRepositoryIDs)

	return nil
}

func resourceGithubCodespaceOrganizationSecretRepositoriesDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	selectedRepositoryIDs := github.SelectedRepoIDs{}
	_, err = client.Codespaces.SetSelectedReposForOrgSecret(ctx, owner, d.Id(), selectedRepositoryIDs)
	if err != nil {
		return err
	}

	return nil
}
