package github

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubActionsEnvironmentVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsEnvironmentVariableCreate,
		Read:   resourceGithubActionsEnvironmentVariableRead,
		Update: resourceGithubActionsEnvironmentVariableUpdate,
		Delete: resourceGithubActionsEnvironmentVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the repository.",
			},
			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the environment.",
			},
			"variable_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Name of the variable.",
				ValidateDiagFunc: validateSecretNameFunc,
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Value of the variable.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of 'actions_variable' creation.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of 'actions_variable' update.",
			},
		},
	}
}

func resourceGithubActionsEnvironmentVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repoName := d.Get("repository").(string)
	envName := d.Get("environment").(string)
	escapedEnvName := url.PathEscape(envName)
	name := d.Get("variable_name").(string)

	variable := &github.ActionsVariable{
		Name:  name,
		Value: d.Get("value").(string),
	}

	_, err := client.Actions.CreateEnvVariable(ctx, owner, repoName, escapedEnvName, variable)
	if err != nil {
		return err
	}

	d.SetId(buildThreePartID(repoName, envName, name))
	return resourceGithubActionsEnvironmentVariableRead(d, meta)
}

func resourceGithubActionsEnvironmentVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repoName := d.Get("repository").(string)
	envName := d.Get("environment").(string)
	escapedEnvName := url.PathEscape(envName)
	name := d.Get("variable_name").(string)

	variable := &github.ActionsVariable{
		Name:  name,
		Value: d.Get("value").(string),
	}

	_, err := client.Actions.UpdateEnvVariable(ctx, owner, repoName, escapedEnvName, variable)
	if err != nil {
		return err
	}

	d.SetId(buildThreePartID(repoName, envName, name))
	return resourceGithubActionsEnvironmentVariableRead(d, meta)
}

func resourceGithubActionsEnvironmentVariableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	repoName, envName, name, err := parseThreePartID(d.Id(), "repository", "environment", "variable_name")
	if err != nil {
		return err
	}
	escapedEnvName := url.PathEscape(envName)

	variable, _, err := client.Actions.GetEnvVariable(ctx, owner, repoName, escapedEnvName, name)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing actions variable %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("repository", repoName)
	d.Set("environment", envName)
	d.Set("variable_name", name)
	d.Set("value", variable.Value)
	d.Set("created_at", variable.CreatedAt.String())
	d.Set("updated_at", variable.UpdatedAt.String())

	return nil
}

func resourceGithubActionsEnvironmentVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	repoName, envName, name, err := parseThreePartID(d.Id(), "repository", "environment", "variable_name")
	if err != nil {
		return err
	}
	escapedEnvName := url.PathEscape(envName)

	_, err = client.Actions.DeleteEnvVariable(ctx, owner, repoName, escapedEnvName, name)

	return err
}
