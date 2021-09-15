package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

	log.Printf("[DEBUG] Reading GitHub branch reference %s/%s (%s)", orgName, repoName, branchRefName)
	ref, resp, err := client.Git.GetRef(context.TODO(), orgName, repoName, branchRefName)
	if err != nil {
		if err, ok := err.(*github.ErrorResponse); ok {
			if err.Response.StatusCode == http.StatusNotFound {
				log.Printf("Error reading GitHub branch reference %s/%s (%s): %s", orgName, repoName, branchRefName, err)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId(buildTwoPartID(repoName, branchName))
	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("ref", *ref.Ref)
	d.Set("sha", *ref.Object.SHA)

	return nil
}
