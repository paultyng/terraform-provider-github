package github

import (
	"context"
	"errors"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubActionsEnterprisePermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsEnterprisePermissionsCreateOrUpdate,
		Read:   resourceGithubActionsEnterprisePermissionsRead,
		Update: resourceGithubActionsEnterprisePermissionsCreateOrUpdate,
		Delete: resourceGithubActionsEnterprisePermissionsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enterprise_slug": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The slug of the enterprise.",
			},
			"allowed_actions": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The permissions policy that controls the actions that are allowed to run. Can be one of: 'all', 'local_only', or 'selected'.",
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"all", "local_only", "selected"}, false), "allowed_actions"),
			},
			"enabled_organizations": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The policy that controls the organizations in the enterprise that are allowed to run GitHub Actions. Can be one of: 'all', 'none', or 'selected'.",
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"all", "none", "selected"}, false), "enabled_organizations"),
			},
			"allowed_actions_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Sets the actions that are allowed in an enterprise. Only available when 'allowed_actions' = 'selected'",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"github_owned_allowed": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether GitHub-owned actions are allowed in the enterprise.",
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
			"enabled_organizations_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Sets the list of selected organizations that are enabled for GitHub Actions in an enterprise. Only available when 'enabled_organizations' = 'selected'.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organization_ids": {
							Type:        schema.TypeSet,
							Description: "List of organization IDs to enable for GitHub Actions.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceGithubActionsEnterpriseAllowedObject(d *schema.ResourceData) (*github.ActionsAllowed, error) {
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
		return &github.ActionsAllowed{},
			errors.New("the allowed_actions_config {} block must be specified if allowed_actions == 'selected'")
	}

	return allowed, nil
}

func resourceGithubActionsEnabledOrganizationsObject(d *schema.ResourceData) ([]int64, error) {
	var enabled []int64

	config := d.Get("enabled_organizations_config").([]interface{})
	if len(config) > 0 {
		data := config[0].(map[string]interface{})
		switch x := data["organization_ids"].(type) {
		case *schema.Set:
			for _, value := range x.List() {
				enabled = append(enabled, int64(value.(int)))
			}
		}
	} else {
		return nil, errors.New("the enabled_organizations_config {} block must be specified if enabled_organizations == 'selected'")
	}
	return enabled, nil
}

func resourceGithubActionsEnterprisePermissionsCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	enterpriseId := d.Get("enterprise_slug").(string)
	allowedActions := d.Get("allowed_actions").(string)
	enabledOrganizations := d.Get("enabled_organizations").(string)

	_, _, err := client.Actions.EditActionsPermissionsInEnterprise(ctx,
		enterpriseId,
		github.ActionsPermissionsEnterprise{
			AllowedActions:       &allowedActions,
			EnabledOrganizations: &enabledOrganizations,
		})
	if err != nil {
		return err
	}

	if allowedActions == "selected" {
		actionsAllowedData, err := resourceGithubActionsEnterpriseAllowedObject(d)
		if err != nil {
			return err
		}
		_, _, err = client.Actions.EditActionsAllowedInEnterprise(ctx,
			enterpriseId,
			*actionsAllowedData)
		if err != nil {
			return err
		}
	}

	if enabledOrganizations == "selected" {
		enabledOrgsData, err := resourceGithubActionsEnabledOrganizationsObject(d)
		if err != nil {
			return err
		}
		_, err = client.Actions.SetEnabledOrgsInEnterprise(ctx,
			enterpriseId,
			enabledOrgsData)
		if err != nil {
			return err
		}
	}

	d.SetId(enterpriseId)
	return resourceGithubActionsEnterprisePermissionsRead(d, meta)
}

func resourceGithubActionsEnterprisePermissionsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()

	actionsPermissions, _, err := client.Actions.GetActionsPermissionsInEnterprise(ctx, d.Id())
	if err != nil {
		return err
	}

	if actionsPermissions.GetAllowedActions() == "selected" {
		actionsAllowed, _, err := client.Actions.GetActionsAllowedInEnterprise(ctx, d.Id())
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

	if actionsPermissions.GetEnabledOrganizations() == "selected" {
		opts := github.ListOptions{PerPage: 10, Page: 1}
		var orgList []int64
		var allOrgs []*github.Organization

		for {
			enabledOrgs, resp, err := client.Actions.ListEnabledOrgsInEnterprise(ctx, d.Id(), &opts)
			if err != nil {
				return err
			}
			allOrgs = append(allOrgs, enabledOrgs.Organizations...)

			opts.Page = resp.NextPage

			if resp.NextPage == 0 {
				break
			}
		}
		for index := range allOrgs {
			orgList = append(orgList, *allOrgs[index].ID)
		}
		if allOrgs != nil {
			if err = d.Set("enabled_organizations_config", []interface{}{
				map[string]interface{}{
					"organization_ids": orgList,
				},
			}); err != nil {
				return err
			}
		} else {
			if err = d.Set("enabled_organizations_config", []interface{}{}); err != nil {
				return err
			}
		}
	}

	if err = d.Set("allowed_actions", actionsPermissions.GetAllowedActions()); err != nil {
		return err
	}
	if err = d.Set("enabled_organizations", actionsPermissions.GetEnabledOrganizations()); err != nil {
		return err
	}
	if err = d.Set("enterprise_slug", d.Id()); err != nil {
		return err
	}

	return nil
}

func resourceGithubActionsEnterprisePermissionsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	// This will nullify any allowedActions elements
	_, _, err := client.Actions.EditActionsPermissionsInEnterprise(ctx,
		d.Get("enterprise_slug").(string),
		github.ActionsPermissionsEnterprise{
			AllowedActions:       github.String("all"),
			EnabledOrganizations: github.String("all"),
		})
	if err != nil {
		return err
	}

	return nil
}
