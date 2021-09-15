package github

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceGithubRelease() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubReleaseRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"retrieve_by": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"latest",
					"id",
					"tag",
				}, false),
			},
			"release_tag": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"release_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"target_commitish": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"draft": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"prerelease": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"published_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asserts_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"upload_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zipball_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tarball_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceGithubReleaseRead(d *schema.ResourceData, meta interface{}) error {
	repository := d.Get("repository").(string)
	owner := d.Get("owner").(string)

	client := meta.(*Owner).v3client
	ctx := context.Background()

	var err error
	var release *github.RepositoryRelease

	switch retrieveBy := strings.ToLower(d.Get("retrieve_by").(string)); retrieveBy {
	case "latest":
		log.Printf("[INFO] Refreshing GitHub latest release from repository %s", repository)
		release, _, err = client.Repositories.GetLatestRelease(ctx, owner, repository)
	case "id":
		releaseID := int64(d.Get("release_id").(int))
		if releaseID == 0 {
			return fmt.Errorf("`release_id` must be set when `retrieve_by` = `id`")
		}

		log.Printf("[INFO] Refreshing GitHub release by id %d from repository %s", releaseID, repository)
		release, _, err = client.Repositories.GetRelease(ctx, owner, repository, releaseID)
	case "tag":
		tag := d.Get("release_tag").(string)
		if tag == "" {
			return fmt.Errorf("`release_tag` must be set when `retrieve_by` = `tag`")
		}

		log.Printf("[INFO] Refreshing GitHub release by tag %s from repository %s", tag, repository)
		release, _, err = client.Repositories.GetReleaseByTag(ctx, owner, repository, tag)
	default:
		return fmt.Errorf("one of: `latest`, `id`, `tag` must be set for `retrieve_by`")
	}

	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(release.GetID(), 10))
	d.Set("release_tag", release.GetTagName())
	d.Set("target_commitish", release.GetTargetCommitish())
	d.Set("name", release.GetName())
	d.Set("body", release.GetBody())
	d.Set("draft", release.GetDraft())
	d.Set("prerelease", release.GetPrerelease())
	d.Set("created_at", release.GetCreatedAt())
	d.Set("published_at", release.GetPublishedAt())
	d.Set("url", release.GetURL())
	d.Set("html_url", release.GetHTMLURL())
	d.Set("asserts_url", release.GetAssetsURL())
	d.Set("upload_url", release.GetUploadURL())
	d.Set("zipball_url", release.GetZipballURL())
	d.Set("tarball_url", release.GetTarballURL())

	return nil
}
