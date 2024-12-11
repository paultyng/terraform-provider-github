package github

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRepositoryDependabotSecurityUpdates() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryDependabotSecurityUpdatesCreateOrUpdate,
		Read:   resourceGithubRepositoryDependabotSecurityUpdatesRead,
		Update: resourceGithubRepositoryDependabotSecurityUpdatesCreateOrUpdate,
		Delete: resourceGithubRepositoryDependabotSecurityUpdatesDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubRepositoryDependabotSecurityUpdatesImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub repository.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "The state of the automated security fixes.",
			},
		},
	}
}

func resourceGithubRepositoryDependabotSecurityUpdatesCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	enabled := d.Get("enabled").(bool)

	ctx := context.Background()
	var err error
	if enabled {
		_, err = client.Repositories.EnableAutomatedSecurityFixes(ctx, owner, repoName)
	} else {
		_, err = client.Repositories.DisableAutomatedSecurityFixes(ctx, owner, repoName)
	}

	if err != nil {
		return err
	}
	d.SetId(repoName)
	return resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta)
}

func resourceGithubRepositoryDependabotSecurityUpdatesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	ctx := context.Background()

	p, _, err := client.Repositories.GetAutomatedSecurityFixes(ctx, orgName, repoName)
	if err != nil {
		return err
	}
	d.Set("enabled", p.Enabled)

	return nil
}

func resourceGithubRepositoryDependabotSecurityUpdatesDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	ctx := context.Background()

	_, err := client.Repositories.DisableAutomatedSecurityFixes(ctx, orgName, repoName)
	if err != nil {
		return err
	}

	return resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta)
}

func resourceGithubRepositoryDependabotSecurityUpdatesImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	repoName := d.Id()

	d.Set("repository", repoName)

	err := resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
