package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubIssueLabels() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubIssueLabelsRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"labels": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"color": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubIssueLabelsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repository := d.Get("repository").(string)

	ctx := context.Background()
	opts := &github.ListOptions{
		PerPage: maxPerPage,
	}

	d.SetId(repository)

	allLabels := make([]interface{}, 0)
	for {
		labels, resp, err := client.Issues.ListLabels(ctx, owner, repository, opts)
		if err != nil {
			return err
		}

		result, err := flattenLabels(labels)
		if err != nil {
			return fmt.Errorf("unable to flatten GitHub Labels (Owner: %q/Repository: %q) : %+v", owner, repository, err)
		}

		allLabels = append(allLabels, result...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	err := d.Set("labels", allLabels)
	if err != nil {
		return err
	}

	return nil
}

func flattenLabels(labels []*github.Label) ([]interface{}, error) {
	if labels == nil {
		return make([]interface{}, 0), nil
	}

	results := make([]interface{}, 0)

	for _, l := range labels {
		result := make(map[string]interface{})

		result["name"] = l.GetName()
		result["color"] = l.GetColor()
		result["description"] = l.GetDescription()
		result["url"] = l.GetURL()

		results = append(results, result)
	}

	return results, nil
}
