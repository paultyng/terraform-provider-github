package github

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRepositoryEnvironmentDeploymentCustomProtectionRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryEnvironmentDeploymentCustomProtectionRuleCreate,
		Read:   resourceGithubRepositoryEnvironmentDeploymentCustomProtectionRuleRead,
		Delete: resourceGithubRepositoryEnvironmentDeploymentCustomProtectionRuleDelete,
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
			"integration_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the custom app that will be enabled on the environment.",
			},
		},
	}

}

func resourceGithubRepositoryEnvironmentDeploymentCustomProtectionRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	envName := d.Get("environment").(string)
	integrationID := d.Get("integration_id").(int)
	escapedEnvName := url.PathEscape(envName)

	createData := github.CustomDeploymentProtectionRuleRequest{
		IntegrationID: github.Int64(int64(integrationID)),
	}

	resultKey, _, err := client.Repositories.CreateCustomDeploymentProtectionRule(ctx, owner, repoName, escapedEnvName, &createData)
	if err != nil {
		return err
	}

	d.SetId(buildThreePartID(repoName, escapedEnvName, strconv.FormatInt(resultKey.GetID(), 10)))
	return resourceGithubRepositoryEnvironmentDeploymentCustomProtectionRuleRead(d, meta)
}

func resourceGithubRepositoryEnvironmentDeploymentCustomProtectionRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	owner := meta.(*Owner).name
	repoName, envName, protectionRuleIdString, err := parseThreePartID(d.Id(), "repository", "environment", "protectionRuleId")
	if err != nil {
		return err
	}

	protectionRuleId, err := strconv.ParseInt(protectionRuleIdString, 10, 64)
	if err != nil {
		return err
	}
	protectionRule, _, err := client.Repositories.GetCustomDeploymentProtectionRule(ctx, owner, repoName, envName, protectionRuleId)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing custom protection for %s/%s/%s from state because it no longer exists in GitHub",
					owner, repoName, envName)
				d.SetId("")
				return nil
			}
		}
		return err
	}
	log.Printf("[INFO] Custom protection rule with node_id %s is enabled", *protectionRule.NodeID)
	return nil
}

func resourceGithubRepositoryEnvironmentDeploymentCustomProtectionRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	owner := meta.(*Owner).name
	repoName, envName, protectionRuleIdString, err := parseThreePartID(d.Id(), "repository", "environment", "protectionRuleId")
	if err != nil {
		return err
	}

	protectionRuleId, err := strconv.ParseInt(protectionRuleIdString, 10, 64)
	if err != nil {
		return err
	}

	_, err = client.Repositories.DisableCustomDeploymentProtectionRule(ctx, owner, repoName, envName, protectionRuleId)
	if err != nil {
		return err
	}

	return nil
}
