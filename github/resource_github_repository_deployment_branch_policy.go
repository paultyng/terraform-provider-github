package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRepositoryDeploymentBranchPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryDeploymentBranchPolicyCreate,
		Read:   resourceGithubRepositoryDeploymentBranchPolicyRead,
		Update: resourceGithubRepositoryDeploymentBranchPolicyUpdate,
		Delete: resourceGithubRepositoryDeploymentBranchPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubRepositoryDeploymentBranchPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub repository name.",
			},
			"environment_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The target environment name.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the branch",
			},
			"etag": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An etag representing the Branch object.",
			},
		},
	}
}

func resourceGithubRepositoryDeploymentBranchPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	environmentName := d.Get("environment_name").(string)
	name := d.Get("name").(string)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	_, _, err = client.Repositories.UpdateDeploymentBranchPolicy(ctx, owner, repoName, environmentName, int64(id), &github.DeploymentBranchPolicyRequest{Name: &name})
	if err != nil {
		return err
	}

	return resourceGithubRepositoryDeploymentBranchPolicyRead(d, meta)
}

func resourceGithubRepositoryDeploymentBranchPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	environmentName := d.Get("environment_name").(string)
	name := d.Get("name").(string)

	policy, _, err := client.Repositories.CreateDeploymentBranchPolicy(ctx, owner, repoName, environmentName, &github.DeploymentBranchPolicyRequest{Name: &name})
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(*policy.ID, 10))

	return resourceGithubRepositoryDeploymentBranchPolicyRead(d, meta)
}

func resourceGithubRepositoryDeploymentBranchPolicyRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	environmentName := d.Get("environment_name").(string)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	policy, resp, err := client.Repositories.GetDeploymentBranchPolicy(ctx, owner, repoName, environmentName, int64(id))
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing deployment branch policy for environment %s: %s from state because it no longer exists in GitHub",
					repoName, environmentName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("repository", repoName); err != nil {
		return err
	}
	if err = d.Set("environment_name", environmentName); err != nil {
		return err
	}
	if err = d.Set("name", policy.Name); err != nil {
		return err
	}

	return nil
}

func resourceGithubRepositoryDeploymentBranchPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	environmentName := d.Get("environment_name").(string)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	_, error := client.Repositories.DeleteDeploymentBranchPolicy(ctx, owner, repoName, environmentName, int64(id))
	if error != nil {
		return error
	}
	return nil
}

func resourceGithubRepositoryDeploymentBranchPolicyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	repoName, environmentName, id, err := parseThreePartID(d.Id(), "repository", "environment_name", "id")
	if err != nil {
		return nil, err
	}

	d.SetId(id)
	if err = d.Set("repository", repoName); err != nil {
		return nil, err
	}
	if err = d.Set("environment_name", environmentName); err != nil {
		return nil, err
	}

	err = resourceGithubRepositoryDeploymentBranchPolicyRead(d, meta)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
