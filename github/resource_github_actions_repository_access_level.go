package github

import (
	"context"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubActionsRepositoryAccessLevel() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsRepositoryAccessLevelCreateOrUpdate,
		Read:   resourceGithubActionsRepositoryAccessLevelRead,
		Update: resourceGithubActionsRepositoryAccessLevelCreateOrUpdate,
		Delete: resourceGithubActionsRepositoryAccessLevelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_level": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Where the actions or reusable workflows of the repository may be used. Possible values are 'none', 'user', 'organization', or 'enterprise'.",
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"none", "user", "organization", "enterprise"}, false), "access_level"),
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

func resourceGithubActionsRepositoryAccessLevelCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	accessLevel := d.Get("access_level").(string)
	actionAccessLevel := github.RepositoryActionsAccessLevel{
		AccessLevel: github.String(accessLevel),
	}

	_, err := client.Repositories.EditActionsAccessLevel(ctx, owner, repoName, actionAccessLevel)
	if err != nil {
		return err
	}

	d.SetId(repoName)
	return resourceGithubActionsRepositoryAccessLevelRead(d, meta)
}

func resourceGithubActionsRepositoryAccessLevelRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Id()
	ctx := context.WithValue(context.Background(), ctxId, repoName)

	actionAccessLevel, _, err := client.Repositories.GetActionsAccessLevel(ctx, owner, repoName)
	if err != nil {
		return err
	}

	_ = d.Set("access_level", actionAccessLevel.GetAccessLevel())

	return nil
}

func resourceGithubActionsRepositoryAccessLevelDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Id()
	ctx := context.WithValue(context.Background(), ctxId, repoName)

	actionAccessLevel := github.RepositoryActionsAccessLevel{
		AccessLevel: github.String("none"),
	}
	_, err := client.Repositories.EditActionsAccessLevel(ctx, owner, repoName, actionAccessLevel)

	return err
}
