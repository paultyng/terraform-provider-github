package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubActionsEnterpriseRunnerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsEnterpriseRunnerGroupCreate,
		Read:   resourceGithubActionsEnterpriseRunnerGroupRead,
		Update: resourceGithubActionsEnterpriseRunnerGroupUpdate,
		Delete: resourceGithubActionsEnterpriseRunnerGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubActionsEnterpriseRunnerGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"enterprise_slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the enterprise.",
			},
			"allows_public_repositories": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether public repositories can be added to the runner group.",
			},
			"default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this is the default runner group.",
			},
			"etag": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An etag representing the runner group object",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the runner group.",
			},
			"runners_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The GitHub API URL for the runner group's runners.",
			},
			"visibility": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The visibility of the runner group.",
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"all", "selected"}, false), "visibility"),
			},
			"restricted_to_workflows": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If 'true', the runner group will be restricted to running only the workflows specified in the 'selected_workflows' array. Defaults to 'false'.",
			},
			"selected_workflows": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of workflows the runner group should be allowed to run. This setting will be ignored unless restricted_to_workflows is set to 'true'.",
			},
			"selected_organization_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Set:         schema.HashInt,
				Optional:    true,
				Description: "List of organization IDs that can access the runner group.",
			},
			"selected_organizations_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "GitHub API URL for the runner group's organizations.",
			},
		},
	}
}

func resourceGithubActionsEnterpriseRunnerGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	name := d.Get("name").(string)
	enterpriseSlug := d.Get("enterprise_slug").(string)
	restrictedToWorkflows := d.Get("restricted_to_workflows").(bool)
	visibility := d.Get("visibility").(string)
	selectedOrganizations, hasSelectedOrganizations := d.GetOk("selected_organization_ids")
	allowsPublicRepositories := d.Get("allows_public_repositories").(bool)

	selectedWorkflows := []string{}
	if workflows, ok := d.GetOk("selected_workflows"); ok {
		for _, workflow := range workflows.([]interface{}) {
			selectedWorkflows = append(selectedWorkflows, workflow.(string))
		}
	}

	if visibility != "selected" && hasSelectedOrganizations {
		return fmt.Errorf("cannot use selected_organization_ids without visibility being set to selected")
	}

	selectedOrganizationIDs := []int64{}

	if hasSelectedOrganizations {
		ids := selectedOrganizations.(*schema.Set).List()

		for _, id := range ids {
			selectedOrganizationIDs = append(selectedOrganizationIDs, int64(id.(int)))
		}
	}

	ctx := context.Background()

	enterpriseRunnerGroup, resp, err := client.Enterprise.CreateEnterpriseRunnerGroup(ctx,
		enterpriseSlug,
		github.CreateEnterpriseRunnerGroupRequest{
			Name:                     &name,
			Visibility:               &visibility,
			SelectedOrganizationIDs:  selectedOrganizationIDs,
			AllowsPublicRepositories: &allowsPublicRepositories,
			RestrictedToWorkflows:    &restrictedToWorkflows,
			SelectedWorkflows:        selectedWorkflows,
		},
	)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(enterpriseRunnerGroup.GetID(), 10))
	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("allows_public_repositories", enterpriseRunnerGroup.GetAllowsPublicRepositories()); err != nil {
		return err
	}
	if err = d.Set("default", enterpriseRunnerGroup.GetDefault()); err != nil {
		return err
	}
	if err = d.Set("name", enterpriseRunnerGroup.GetName()); err != nil {
		return err
	}
	if err = d.Set("runners_url", enterpriseRunnerGroup.GetRunnersURL()); err != nil {
		return err
	}
	if err = d.Set("selected_organizations_url", enterpriseRunnerGroup.GetSelectedOrganizationsURL()); err != nil {
		return err
	}
	if err = d.Set("visibility", enterpriseRunnerGroup.GetVisibility()); err != nil {
		return err
	}
	if err = d.Set("selected_organization_ids", selectedOrganizationIDs); err != nil { // Note: enterpriseRunnerGroup has no method to get selected organization IDs
		return err
	}
	if err = d.Set("restricted_to_workflows", enterpriseRunnerGroup.GetRestrictedToWorkflows()); err != nil {
		return err
	}
	if err = d.Set("selected_workflows", enterpriseRunnerGroup.SelectedWorkflows); err != nil {
		return err
	}
	if err = d.Set("enterprise_slug", enterpriseSlug); err != nil {
		return err
	}

	return resourceGithubActionsEnterpriseRunnerGroupRead(d, meta)
}

func getEnterpriseRunnerGroup(client *github.Client, ctx context.Context, ent string, groupID int64) (*github.EnterpriseRunnerGroup, *github.Response, error) {
	enterpriseRunnerGroup, resp, err := client.Enterprise.GetEnterpriseRunnerGroup(ctx, ent, groupID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == http.StatusNotModified {
			// ignore error StatusNotModified
			return enterpriseRunnerGroup, resp, nil
		}
	}
	return enterpriseRunnerGroup, resp, err
}

func resourceGithubActionsEnterpriseRunnerGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	enterpriseSlug := d.Get("enterprise_slug").(string)
	runnerGroupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	enterpriseRunnerGroup, resp, err := getEnterpriseRunnerGroup(client, ctx, enterpriseSlug, runnerGroupID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing enterprise runner group %s/%s from state because it no longer exists in GitHub",
					enterpriseSlug, d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	//if runner group is nil (typically not modified) we can return early
	if enterpriseRunnerGroup == nil {
		return nil
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("allows_public_repositories", enterpriseRunnerGroup.GetAllowsPublicRepositories()); err != nil {
		return err
	}
	if err = d.Set("default", enterpriseRunnerGroup.GetDefault()); err != nil {
		return err
	}
	if err = d.Set("name", enterpriseRunnerGroup.GetName()); err != nil {
		return err
	}
	if err = d.Set("runners_url", enterpriseRunnerGroup.GetRunnersURL()); err != nil {
		return err
	}
	if err = d.Set("selected_organizations_url", enterpriseRunnerGroup.GetSelectedOrganizationsURL()); err != nil {
		return err
	}
	if err = d.Set("visibility", enterpriseRunnerGroup.GetVisibility()); err != nil {
		return err
	}
	if err = d.Set("restricted_to_workflows", enterpriseRunnerGroup.GetRestrictedToWorkflows()); err != nil {
		return err
	}
	if err = d.Set("selected_workflows", enterpriseRunnerGroup.SelectedWorkflows); err != nil {
		return err
	}
	if err = d.Set("enterprise_slug", enterpriseSlug); err != nil {
		return err
	}

	selectedOrganizationIDs := []int64{}
	optionsOrgs := github.ListOptions{
		PerPage: maxPerPage,
	}

	for {
		enterpriseRunnerGroupOrganizations, resp, err := client.Enterprise.ListOrganizationAccessRunnerGroup(ctx, enterpriseSlug, runnerGroupID, &optionsOrgs)
		if err != nil {
			return err
		}

		for _, org := range enterpriseRunnerGroupOrganizations.Organizations {
			selectedOrganizationIDs = append(selectedOrganizationIDs, *org.ID)
		}

		if resp.NextPage == 0 {
			break
		}

		optionsOrgs.Page = resp.NextPage
	}

	if err = d.Set("selected_organization_ids", selectedOrganizationIDs); err != nil {
		return err
	}

	return nil
}

func resourceGithubActionsEnterpriseRunnerGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	name := d.Get("name").(string)
	enterpriseSlug := d.Get("enterprise_slug").(string)
	visibility := d.Get("visibility").(string)
	restrictedToWorkflows := d.Get("restricted_to_workflows").(bool)
	selectedWorkflows := []string{}
	allowsPublicRepositories := d.Get("allows_public_repositories").(bool)
	if workflows, ok := d.GetOk("selected_workflows"); ok {
		for _, workflow := range workflows.([]interface{}) {
			selectedWorkflows = append(selectedWorkflows, workflow.(string))
		}
	}

	options := github.UpdateEnterpriseRunnerGroupRequest{
		Name:                     &name,
		Visibility:               &visibility,
		RestrictedToWorkflows:    &restrictedToWorkflows,
		SelectedWorkflows:        selectedWorkflows,
		AllowsPublicRepositories: &allowsPublicRepositories,
	}

	runnerGroupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	if _, _, err := client.Enterprise.UpdateEnterpriseRunnerGroup(ctx, enterpriseSlug, runnerGroupID, options); err != nil {
		return err
	}

	selectedOrganizations, hasSelectedOrganizations := d.GetOk("selected_organization_ids")
	selectedOrganizationIDs := []int64{}

	if hasSelectedOrganizations {
		ids := selectedOrganizations.(*schema.Set).List()

		for _, id := range ids {
			selectedOrganizationIDs = append(selectedOrganizationIDs, int64(id.(int)))
		}
	}

	orgOptions := github.SetOrgAccessRunnerGroupRequest{SelectedOrganizationIDs: selectedOrganizationIDs}

	if _, err := client.Enterprise.SetOrganizationAccessRunnerGroup(ctx, enterpriseSlug, runnerGroupID, orgOptions); err != nil {
		return err
	}

	return resourceGithubActionsEnterpriseRunnerGroupRead(d, meta)
}

func resourceGithubActionsEnterpriseRunnerGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	enterpriseSlug := d.Get("enterprise_slug").(string)
	enterpriseRunnerGroupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[INFO] Deleting enterprise runner group: %s/%s (%s)", enterpriseSlug, d.Get("name"), d.Id())
	_, err = client.Enterprise.DeleteEnterpriseRunnerGroup(ctx, enterpriseSlug, enterpriseRunnerGroupID)
	return err
}

func resourceGithubActionsEnterpriseRunnerGroupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import specified: supplied import must be written as <enterprise_slug>/<runner_group_id>")
	}

	enterpriseId, runnerGroupID := parts[0], parts[1]

	d.SetId(runnerGroupID)
	d.Set("enterprise_slug", enterpriseId)

	return []*schema.ResourceData{d}, nil
}
