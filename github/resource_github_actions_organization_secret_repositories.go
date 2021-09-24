package github

import (
	"context"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubActionsOrganizationSecretRepositories() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsOrganizationSecretRepositoriesCreateOrUpdate,
		Read:   resourceGithubActionsOrganizationSecretRepositoriesRead,
		Update: resourceGithubActionsOrganizationSecretRepositoriesCreateOrUpdate,
		Delete: resourceGithubActionsOrganizationSecretRepositoriesDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"secret_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateSecretNameFunc,
			},
			"selected_repository_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Set:      schema.HashInt,
				Required: true,
			},
		},
	}
}

func resourceGithubActionsOrganizationSecretRepositoriesCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
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

	_, err = client.Actions.SetSelectedReposForOrgSecret(ctx, owner, secretName, selectedRepositoryIDs)
	if err != nil {
		return err
	}

	d.SetId(secretName)
	return resourceGithubActionsOrganizationSecretRepositoriesRead(d, meta)
}

func resourceGithubActionsOrganizationSecretRepositoriesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	selectedRepositoryIDs := []int64{}
	opt := &github.ListOptions{
		PerPage: maxPerPage,
	}
	for {
		results, resp, err := client.Actions.ListSelectedReposForOrgSecret(ctx, owner, d.Id(), opt)
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

func resourceGithubActionsOrganizationSecretRepositoriesDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	selectedRepositoryIDs := []int64{}
	_, err = client.Actions.SetSelectedReposForOrgSecret(ctx, owner, d.Id(), selectedRepositoryIDs)
	if err != nil {
		return err
	}

	return nil
}
