package github

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRepositoryEnvironmentDeploymentPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryEnvironmentDeploymentPolicyCreate,
		Read:   resourceGithubRepositoryEnvironmentDeploymentPolicyRead,
		Update: resourceGithubRepositoryEnvironmentDeploymentPolicyUpdate,
		Delete: resourceGithubRepositoryEnvironmentDeploymentPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the repository. The name is not case sensitive.",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the environment.",
			},
			"branch_pattern": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The name pattern that branches must match in order to deploy to the environment.",
			},
		},
	}

}

func resourceGithubRepositoryEnvironmentDeploymentPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	envName := d.Get("environment").(string)
	branchPattern := d.Get("branch_pattern").(string)
	escapedEnvName := url.PathEscape(envName)

	createData := github.DeploymentBranchPolicyRequest{
		Name: github.String(branchPattern),
	}

	resultKey, _, err := client.Repositories.CreateDeploymentBranchPolicy(ctx, owner, repoName, escapedEnvName, &createData)
	if err != nil {
		return err
	}

	d.SetId(buildThreePartID(repoName, escapedEnvName, strconv.FormatInt(resultKey.GetID(), 10)))
	return resourceGithubRepositoryEnvironmentDeploymentPolicyRead(d, meta)
}

func resourceGithubRepositoryEnvironmentDeploymentPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	owner := meta.(*Owner).name
	repoName, envName, branchPolicyIdString, err := parseThreePartID(d.Id(), "repository", "environment", "branchPolicyId")
	if err != nil {
		return err
	}

	branchPolicyId, err := strconv.ParseInt(branchPolicyIdString, 10, 64)
	if err != nil {
		return err
	}

	branchPolicy, _, err := client.Repositories.GetDeploymentBranchPolicy(ctx, owner, repoName, envName, branchPolicyId)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing branch deployment policy for %s/%s/%s from state because it no longer exists in GitHub",
					owner, repoName, envName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("branch_pattern", branchPolicy.GetName())
	return nil
}

func resourceGithubRepositoryEnvironmentDeploymentPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	envName := d.Get("environment").(string)
	branchPattern := d.Get("branch_pattern").(string)
	escapedEnvName := url.PathEscape(envName)
	_, _, branchPolicyIdString, err := parseThreePartID(d.Id(), "repository", "environment", "branchPolicyId")
	if err != nil {
		return err
	}

	branchPolicyId, err := strconv.ParseInt(branchPolicyIdString, 10, 64)
	if err != nil {
		return err
	}

	updateData := github.DeploymentBranchPolicyRequest{
		Name: github.String(branchPattern),
	}

	resultKey, _, err := client.Repositories.UpdateDeploymentBranchPolicy(ctx, owner, repoName, escapedEnvName, branchPolicyId, &updateData)
	if err != nil {
		return err
	}
	d.SetId(buildThreePartID(repoName, escapedEnvName, strconv.FormatInt(resultKey.GetID(), 10)))
	return resourceGithubRepositoryEnvironmentDeploymentPolicyRead(d, meta)
}

func resourceGithubRepositoryEnvironmentDeploymentPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	owner := meta.(*Owner).name
	repoName, envName, branchPolicyIdString, err := parseThreePartID(d.Id(), "repository", "environment", "branchPolicyId")
	if err != nil {
		return err
	}

	branchPolicyId, err := strconv.ParseInt(branchPolicyIdString, 10, 64)
	if err != nil {
		return err
	}

	_, err = client.Repositories.DeleteDeploymentBranchPolicy(ctx, owner, repoName, envName, branchPolicyId)
	if err != nil {
		return err
	}

	return nil
}
