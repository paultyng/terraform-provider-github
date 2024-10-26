package github

import (
	"context"
	"errors"
	"log"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubActionsOrganizationPermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsOrganizationPermissionsCreateOrUpdate,
		Read:   resourceGithubActionsOrganizationPermissionsRead,
		Update: resourceGithubActionsOrganizationPermissionsCreateOrUpdate,
		Delete: resourceGithubActionsOrganizationPermissionsDelete,
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
			"enabled_repositories": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The policy that controls the repositories in the organization that are allowed to run GitHub Actions. Can be one of: 'all', 'none', or 'selected'.",
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"all", "none", "selected"}, false), "enabled_repositories"),
			},
			"allowed_actions_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Sets the actions that are allowed in an organization. Only available when 'allowed_actions' = 'selected'",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"github_owned_allowed": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether GitHub-owned actions are allowed in the organization.",
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
			"enabled_repositories_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Sets the list of selected repositories that are enabled for GitHub Actions in an organization. Only available when 'enabled_repositories' = 'selected'.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_ids": {
							Type:        schema.TypeSet,
							Description: "List of repository IDs to enable for GitHub Actions.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceGithubActionsOrganizationAllowedObject(d *schema.ResourceData) (*github.ActionsAllowed, error) {
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

func resourceGithubActionsEnabledRepositoriesObject(d *schema.ResourceData) ([]int64, error) {
	var enabled []int64

	config := d.Get("enabled_repositories_config").([]interface{})
	if len(config) > 0 {
		data := config[0].(map[string]interface{})
		switch x := data["repository_ids"].(type) {
		case *schema.Set:
			for _, value := range x.List() {
				enabled = append(enabled, int64(value.(int)))
			}
		}
	} else {
		return nil, errors.New("the enabled_repositories_config {} block must be specified if enabled_repositories == 'selected'")
	}
	return enabled, nil
}

func resourceGithubActionsOrganizationPermissionsCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	allowedActions := d.Get("allowed_actions").(string)
	enabledRepositories := d.Get("enabled_repositories").(string)

	_, _, err = client.Actions.EditActionsPermissions(ctx,
		orgName,
		github.ActionsPermissions{
			AllowedActions:      &allowedActions,
			EnabledRepositories: &enabledRepositories,
		})
	if err != nil {
		return err
	}

	if allowedActions == "selected" {
		actionsAllowedData, err := resourceGithubActionsOrganizationAllowedObject(d)
		if err != nil {
			return err
		}
		if actionsAllowedData != nil {
			log.Printf("[DEBUG] Allowed actions config is set")
			_, _, err = client.Actions.EditActionsAllowed(ctx,
				orgName,
				*actionsAllowedData)
			if err != nil {
				return err
			}
		} else {
			log.Printf("[DEBUG] Allowed actions config not set, skipping")
		}
	}

	if enabledRepositories == "selected" {
		enabledReposData, err := resourceGithubActionsEnabledRepositoriesObject(d)
		if err != nil {
			return err
		}
		_, err = client.Actions.SetEnabledReposInOrg(ctx,
			orgName,
			enabledReposData)
		if err != nil {
			return err
		}
	}

	d.SetId(orgName)
	return resourceGithubActionsOrganizationPermissionsRead(d, meta)
}

func resourceGithubActionsOrganizationPermissionsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	actionsPermissions, _, err := client.Actions.GetActionsPermissions(ctx, d.Id())
	if err != nil {
		return err
	}

	// only load and fill allowed_actions_config if allowed_actions_config is also set
	// in the TF code. (see #2105)
	// on initial import there might not be any value in the state, then we have to import the data
	// -> but we can only load an existing state if the current config is set to "selected" (see #2182)
	allowedActions := d.Get("allowed_actions").(string)
	allowedActionsConfig := d.Get("allowed_actions_config").([]interface{})

	serverHasAllowedActionsConfig := actionsPermissions.GetAllowedActions() == "selected"
	userWantsAllowedActionsConfig := (allowedActions == "selected" && len(allowedActionsConfig) > 0) || allowedActions == ""

	if serverHasAllowedActionsConfig && userWantsAllowedActionsConfig {
		actionsAllowed, _, err := client.Actions.GetActionsAllowed(ctx, d.Id())
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

	if actionsPermissions.GetEnabledRepositories() == "selected" {
		opts := github.ListOptions{PerPage: 10, Page: 1}
		var repoList []int64
		var allRepos []*github.Repository

		for {
			enabledRepos, resp, err := client.Actions.ListEnabledReposInOrg(ctx, d.Id(), &opts)
			if err != nil {
				return err
			}
			allRepos = append(allRepos, enabledRepos.Repositories...)

			opts.Page = resp.NextPage

			if resp.NextPage == 0 {
				break
			}
		}
		for index := range allRepos {
			repoList = append(repoList, *allRepos[index].ID)
		}
		if allRepos != nil {
			if err = d.Set("enabled_repositories_config", []interface{}{
				map[string]interface{}{
					"repository_ids": repoList,
				},
			}); err != nil {
				return err
			}
		} else {
			if err = d.Set("enabled_repositories_config", []interface{}{}); err != nil {
				return err
			}
		}
	}

	if err = d.Set("allowed_actions", actionsPermissions.GetAllowedActions()); err != nil {
		return err
	}
	if err = d.Set("enabled_repositories", actionsPermissions.GetEnabledRepositories()); err != nil {
		return err
	}

	return nil
}

func resourceGithubActionsOrganizationPermissionsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	// This will nullify any allowedActions elements
	_, _, err = client.Actions.EditActionsPermissions(ctx,
		orgName,
		github.ActionsPermissions{
			AllowedActions:      github.String("all"),
			EnabledRepositories: github.String("all"),
		})
	if err != nil {
		return err
	}

	return nil
}
