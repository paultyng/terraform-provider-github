package github

import (
	"context"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubOrganizationWebhooks() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubOrganizationWebhooksRead,

		Schema: map[string]*schema.Schema{
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

func dataSourceGithubOrganizationWebhooksRead(d *schema.ResourceData, meta interface{}) error {
	owner := meta.(*Owner).name

	client := meta.(*Owner).v3client
	ctx := context.Background()

	options := &github.ListOptions{
		PerPage: 100,
	}

	results := make([]map[string]interface{}, 0)
	for {
		hooks, resp, err := client.Organizations.ListHooks(ctx, owner, options)
		if err != nil {
			return err
		}

		results = append(results, flattenGitHubWebhooks(hooks)...)
		if resp.NextPage == 0 {
			break
		}

		options.Page = resp.NextPage
	}

	d.SetId(owner)
	err := d.Set("webhooks", results)
	if err != nil {
		return err
	}

	return nil
}
