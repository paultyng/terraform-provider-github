package github

import (
	"context"

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

	d.SetId(repository)

	labels, err := listLabels(client, ctx, owner, repository)
	if err != nil {
		return err
	}

	err = d.Set("labels", flattenLabels(labels))
	if err != nil {
		return err
	}

	return nil
}
