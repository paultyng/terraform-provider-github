package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubBranch() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubBranchRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ref": {
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

func dataSourceGithubBranchRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	branchName := d.Get("branch").(string)
	branchRefName := "refs/heads/" + branchName

	ref, resp, err := client.Git.GetRef(context.TODO(), orgName, repoName, branchRefName)
	if err != nil {
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == http.StatusNotFound {
				log.Printf("[DEBUG] Missing GitHub branch %s/%s (%s)", orgName, repoName, branchRefName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(buildTwoPartID(repoName, branchName))
	err = d.Set("etag", resp.Header.Get("ETag"))
	if err != nil {
		return err
	}
	err = d.Set("ref", *ref.Ref)
	if err != nil {
		return err
	}
	err = d.Set("sha", *ref.Object.SHA)
	if err != nil {
		return err
	}

	return nil
}
