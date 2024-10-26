package github

import (
	"context"
	"log"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubActionsRepositoryPermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsRepositoryPermissionsCreateOrUpdate,
		Read:   resourceGithubActionsRepositoryPermissionsRead,
		Update: resourceGithubActionsRepositoryPermissionsCreateOrUpdate,
		Delete: resourceGithubActionsRepositoryPermissionsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allowed_actions": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The permissions policy that controls the actions that are allowed to run. Can be one of: 'all', 'local_only', or 'selected'.",
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"all", "local_only", "selected"}, false), "allowed_actions"),
			},
			"allowed_actions_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Sets the actions that are allowed in an repository. Only available when 'allowed_actions' = 'selected'. ",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"github_owned_allowed": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether GitHub-owned actions are allowed in the repository.",
						},
						"patterns_allowed": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Specifies a list of string-matching patterns to allow specific action(s). Wildcards, tags, and SHAs are allowed. For example, 'monalisa/octocat@', 'monalisa/octocat@v2', 'monalisa/'.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
						"verified_allowed": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether actions in GitHub Marketplace from verified creators are allowed. Set to 'true' to allow all GitHub Marketplace actions by verified creators.",
						},
					},
				},
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Should GitHub actions be enabled on this repository.",
			},
			"repository": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The GitHub repository.",
				ValidateDiagFunc: toDiagFunc(validation.StringLenBetween(1, 100), "repository"),
			},
		},
	}
}

func resourceGithubActionsRepositoryAllowedObject(d *schema.ResourceData) (*github.ActionsAllowed, error) {
	allowed := &github.ActionsAllowed{}

	config := d.Get("allowed_actions_config").([]interface{})
	if len(config) > 0 {
		data := config[0].(map[string]interface{})
		switch x := data["github_owned_allowed"].(type) {
		case bool:
			allowed.GithubOwnedAllowed = &x
		}

		switch x := data["verified_allowed"].(type) {
		case bool:
			allowed.VerifiedAllowed = &x
		}

		patternsAllowed := []string{}

		switch t := data["patterns_allowed"].(type) {
		case *schema.Set:
			for _, value := range t.List() {
				patternsAllowed = append(patternsAllowed, value.(string))
			}
		}

		allowed.PatternsAllowed = patternsAllowed
	} else {
		return nil, nil
	}

	return allowed, nil
}

func resourceGithubActionsRepositoryPermissionsCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	allowedActions := d.Get("allowed_actions").(string)
	enabled := d.Get("enabled").(bool)
	log.Printf("[DEBUG] Actions enabled: %t", enabled)

	repoActionPermissions := github.ActionsPermissionsRepository{
		Enabled: &enabled,
	}

	// Only specify `allowed_actions` if actions are enabled
	if enabled {
		repoActionPermissions.AllowedActions = &allowedActions
	}

	_, _, err := client.Repositories.EditActionsPermissions(ctx,
		owner,
		repoName,
		repoActionPermissions,
	)
	if err != nil {
		return err
	}

	if allowedActions == "selected" {
		actionsAllowedData, err := resourceGithubActionsRepositoryAllowedObject(d)
		if err != nil {
			return err
		}
		if actionsAllowedData != nil {
			log.Printf("[DEBUG] Allowed actions config is set")
			_, _, err = client.Repositories.EditActionsAllowed(ctx,
				owner,
				repoName,
				*actionsAllowedData)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] Allowed actions config not set, skipping")
		}
	}

	d.SetId(repoName)
	return resourceGithubActionsRepositoryPermissionsRead(d, meta)
}

func resourceGithubActionsRepositoryPermissionsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Id()
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	actionsPermissions, _, err := client.Repositories.GetActionsPermissions(ctx, owner, repoName)
	if err != nil {
		return err
	}

	// only load and fill allowed_actions_config if allowed_actions_config is also set
	// in the TF code. (see #2105)
	// on initial import there might not be any value in the state, then we have to import the data
	// -> but we can only load an existing state if the current config is set to "selected" (see #2182)
	allowedActions := d.Get("allowed_actions").(string)
	allowedActionsConfig := d.Get("allowed_actions_config").([]interface{})

	serverHasAllowedActionsConfig := actionsPermissions.GetAllowedActions() == "selected" && actionsPermissions.GetEnabled()
	userWantsAllowedActionsConfig := (allowedActions == "selected" && len(allowedActionsConfig) > 0) || allowedActions == ""

	if serverHasAllowedActionsConfig && userWantsAllowedActionsConfig {
		actionsAllowed, _, err := client.Repositories.GetActionsAllowed(ctx, owner, repoName)
		if err != nil {
			return err
		}

		// If actionsAllowed set to local/all by removing all actions config settings, the response will be empty
		if actionsAllowed != nil {
			if err = d.Set("allowed_actions_config", []interface{}{
				map[string]interface{}{
					"github_owned_allowed": actionsAllowed.GetGithubOwnedAllowed(),
					"patterns_allowed":     actionsAllowed.PatternsAllowed,
					"verified_allowed":     actionsAllowed.GetVerifiedAllowed(),
				},
			}); err != nil {
				return err
			}
		}
	} else {
		if err = d.Set("allowed_actions_config", []interface{}{}); err != nil {
			return err
		}
	}

	if err = d.Set("allowed_actions", actionsPermissions.GetAllowedActions()); err != nil {
		return err
	}
	if err = d.Set("enabled", actionsPermissions.GetEnabled()); err != nil {
		return err
	}
	if err = d.Set("repository", repoName); err != nil {
		return err
	}

	return nil
}

func resourceGithubActionsRepositoryPermissionsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Id()

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	// Reset the repo to "default" settings
	repoActionPermissions := github.ActionsPermissionsRepository{
		AllowedActions: github.String("all"),
		Enabled:        github.Bool(true),
	}

	_, _, err := client.Repositories.EditActionsPermissions(ctx,
		owner,
		repoName,
		repoActionPermissions,
	)
	if err != nil {
		return err
	}

	return nil
}
