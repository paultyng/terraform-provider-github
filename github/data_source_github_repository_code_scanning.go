package github

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubRepositoryCodeScanning() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubRepositoryCodeScanningRead,
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"languages": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"query_suite": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGithubRepositoryCodeScanningRead(d *schema.ResourceData, meta interface{}) error {
	repository := d.Get("repository").(string)
	owner := meta.(*Owner).name

	client := meta.(*Owner).v3client
	ctx := context.Background()

	config, _, err := client.CodeScanning.GetDefaultSetupConfiguration(
		ctx,
		owner,
		repository,
	)
	if err != nil {
		return err
	}

	timeString := ""

	if config.UpdatedAt != nil {
		timeString = config.UpdatedAt.String()
	}

	d.SetId(repository)
	d.Set("languages", config.Languages)
	d.Set("query_suite", config.GetQuerySuite())
	d.Set("state", config.GetState())
	d.Set("updated_at", timeString)

	return nil
}
