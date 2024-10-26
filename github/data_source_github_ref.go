package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubRef() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubRefRead,

		Schema: map[string]*schema.Schema{
			"ref": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sha": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGithubRefRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner, ok := d.Get("owner").(string)
	if !ok {
		owner = meta.(*Owner).name
	}
	repoName := d.Get("repository").(string)
	ref := d.Get("ref").(string)

	refData, resp, err := client.Git.GetRef(context.TODO(), owner, repoName, ref)
	if err != nil {
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == http.StatusNotFound {
				log.Printf("[DEBUG] Missing GitHub ref %s/%s (%s)", owner, repoName, ref)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(buildTwoPartID(repoName, ref))
	err = d.Set("etag", resp.Header.Get("ETag"))
	if err != nil {
		return err
	}
	err = d.Set("sha", *refData.Object.SHA)
	if err != nil {
		return err
	}

	return nil
}
