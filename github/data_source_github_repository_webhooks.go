package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubRepositoryWebhooks() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubRepositoryWebhooksRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"webhooks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubRepositoryWebhooksRead(d *schema.ResourceData, meta interface{}) error {
	repository := d.Get("repository").(string)
	owner := meta.(*Owner).name

	client := meta.(*Owner).v3client
	ctx := context.Background()

	options := &github.ListOptions{
		PerPage: 100,
	}

	results := make([]map[string]interface{}, 0)
	for {
		hooks, resp, err := client.Repositories.ListHooks(ctx, owner, repository, options)
		if err != nil {
			return err
		}

		results = append(results, flattenGitHubWebhooks(hooks)...)

		if resp.NextPage == 0 {
			break
		}

		options.Page = resp.NextPage
	}

	d.SetId(fmt.Sprintf("%s/%s", owner, repository))
	if err := d.Set("repository", repository); err != nil {
		return err
	}
	if err := d.Set("webhooks", results); err != nil {
		return err
	}

	return nil
}

func flattenGitHubWebhooks(hooks []*github.Hook) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)

	if hooks == nil {
		return results
	}

	for _, hook := range hooks {
		result := make(map[string]interface{})

		result["id"] = hook.ID
		result["type"] = hook.Type
		result["name"] = hook.Name
		result["url"] = hook.URL
		result["active"] = hook.Active

		results = append(results, result)
	}

	return results
}
