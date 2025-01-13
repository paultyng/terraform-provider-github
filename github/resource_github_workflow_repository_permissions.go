package github

import (
	"context"

	"github.com/google/go-github/v62/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubWorkflowRepositoryPermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubWorkflowRepositoryPermissionsCreateOrUpdate,
		Read:   resourceGithubWorkflowRepositoryPermissionsRead,
		Update: resourceGithubWorkflowRepositoryPermissionsCreateOrUpdate,
		Delete: resourceGithubWorkflowRepositoryPermissionsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"default_workflow_permissions": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The default workflow permissions granted to the GITHUB_TOKEN when running workflows.",
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"read", "write"}, false), "default_workflow_permissions"),
			},
			"can_approve_pull_request_reviews": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether GitHub Actions can approve pull requests. Enabling this can be a security risk.",
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

func resourceGithubWorkflowRepositoryPermissionsCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	defaultWorkflowPermissions := d.Get("default_workflow_permissions").(string)
	canApprovePullRequestReviews := d.Get("can_approve_pull_request_reviews").(bool)

	repoWorkflowPermissions := github.DefaultWorkflowPermissionRepository{
		DefaultWorkflowPermissions:   &defaultWorkflowPermissions,
		CanApprovePullRequestReviews: &canApprovePullRequestReviews,
	}

	_, _, err := client.Repositories.EditDefaultWorkflowPermissions(ctx,
		owner,
		repoName,
		repoWorkflowPermissions,
	)
	if err != nil {
		return err
	}

	d.SetId(repoName)
	return resourceGithubWorkflowRepositoryPermissionsRead(d, meta)
}

func resourceGithubWorkflowRepositoryPermissionsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	owner := meta.(*Owner).name
	repoName := d.Id()
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	workflowsPermissions, _, err := client.Repositories.GetDefaultWorkflowPermissions(ctx, owner, repoName)
	if err != nil {
		return err
	}

	if err = d.Set("default_workflow_permissions", workflowsPermissions.GetDefaultWorkflowPermissions()); err != nil {
		return err
	}
	if err = d.Set("can_approve_pull_request_reviews", workflowsPermissions.GetCanApprovePullRequestReviews()); err != nil {
		return err
	}
	if err = d.Set("repository", repoName); err != nil {
		return err
	}

	return nil
}

func resourceGithubWorkflowRepositoryPermissionsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Id()

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	// Reset the repo to "default" settings
	repoWorkflowPermissions := github.DefaultWorkflowPermissionRepository{
		DefaultWorkflowPermissions:   github.String("read"),
		CanApprovePullRequestReviews: github.Bool(false),
	}

	_, _, err := client.Repositories.EditDefaultWorkflowPermissions(ctx,
		owner,
		repoName,
		repoWorkflowPermissions,
	)
	if err != nil {
		return err
	}

	return nil
}
