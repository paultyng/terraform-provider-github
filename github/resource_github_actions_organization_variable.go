package github

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubActionsOrganizationVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsOrganizationVariableCreate,
		Read:   resourceGithubActionsOrganizationVariableRead,
		Update: resourceGithubActionsOrganizationVariableUpdate,
		Delete: resourceGithubActionsOrganizationVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"visibility": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateValueFunc([]string{"all", "private", "selected"}),
				ForceNew:         true,
				Description:      "Configures the access that repositories have to the organization variable. Must be one of 'all', 'private', or 'selected'. 'selected_repository_ids' is required if set to 'selected'.",
			},
			"selected_repository_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Set:         schema.HashInt,
				Optional:    true,
				Description: "An array of repository ids that can access the organization variable.",
			},
		},
	}
}

func resourceGithubActionsOrganizationVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	name := d.Get("variable_name").(string)

	visibility := d.Get("visibility").(string)
	selectedRepositories, hasSelectedRepositories := d.GetOk("selected_repository_ids")

	if visibility != "selected" && hasSelectedRepositories {
		return fmt.Errorf("cannot use selected_repository_ids without visibility being set to selected")
	}

	selectedRepositoryIDs := []int64{}

	if hasSelectedRepositories {
		ids := selectedRepositories.(*schema.Set).List()

		for _, id := range ids {
			selectedRepositoryIDs = append(selectedRepositoryIDs, int64(id.(int)))
		}
	}

	repoIDs := github.SelectedRepoIDs(selectedRepositoryIDs)

	variable := &github.ActionsVariable{
		Name:                  name,
		Value:                 d.Get("value").(string),
		Visibility:            &visibility,
		SelectedRepositoryIDs: &repoIDs,
	}
	_, err := client.Actions.CreateOrgVariable(ctx, owner, variable)
	if err != nil {
		return err
	}

	d.SetId(name)
	return resourceGithubActionsOrganizationVariableRead(d, meta)
}

func resourceGithubActionsOrganizationVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	name := d.Get("variable_name").(string)

	visibility := d.Get("visibility").(string)
	selectedRepositories, hasSelectedRepositories := d.GetOk("selected_repository_ids")

	if visibility != "selected" && hasSelectedRepositories {
		return fmt.Errorf("cannot use selected_repository_ids without visibility being set to selected")
	}

	selectedRepositoryIDs := []int64{}

	if hasSelectedRepositories {
		ids := selectedRepositories.(*schema.Set).List()

		for _, id := range ids {
			selectedRepositoryIDs = append(selectedRepositoryIDs, int64(id.(int)))
		}
	}

	repoIDs := github.SelectedRepoIDs(selectedRepositoryIDs)

	variable := &github.ActionsVariable{
		Name:                  name,
		Value:                 d.Get("value").(string),
		Visibility:            &visibility,
		SelectedRepositoryIDs: &repoIDs,
	}

	_, err := client.Actions.UpdateOrgVariable(ctx, owner, variable)
	if err != nil {
		return err
	}

	d.SetId(name)
	return resourceGithubActionsOrganizationVariableRead(d, meta)
}

func resourceGithubActionsOrganizationVariableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	name := d.Id()

	variable, _, err := client.Actions.GetOrgVariable(ctx, owner, name)
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

	if err = d.Set("variable_name", name); err != nil {
		return err
	}
	if err = d.Set("value", variable.Value); err != nil {
		return err
	}
	if err = d.Set("created_at", variable.CreatedAt.String()); err != nil {
		return err
	}
	if err = d.Set("updated_at", variable.UpdatedAt.String()); err != nil {
		return err
	}
	if err = d.Set("visibility", *variable.Visibility); err != nil {
		return err
	}

	selectedRepositoryIDs := []int64{}

	if *variable.Visibility == "selected" {
		opt := &github.ListOptions{
			PerPage: 30,
		}
		for {
			results, resp, err := client.Actions.ListSelectedReposForOrgVariable(ctx, owner, d.Id(), opt)
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
	}

	if err = d.Set("selected_repository_ids", selectedRepositoryIDs); err != nil {
		return err
	}

	return nil
}

func resourceGithubActionsOrganizationVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	name := d.Id()

	_, err := client.Actions.DeleteOrgVariable(ctx, owner, name)

	return err
}
