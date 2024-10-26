package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRelease() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubReleaseCreateUpdate,
		Update: resourceGithubReleaseCreateUpdate,
		Read:   resourceGithubReleaseRead,
		Delete: resourceGithubReleaseDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubReleaseImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the repository.",
			},
			"tag_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the tag.",
			},
			"target_commitish": {
				Type:        schema.TypeString,
				Default:     "main",
				Optional:    true,
				ForceNew:    true,
				Description: " The branch name or commit SHA the tag is created from.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "The name of the release.",
			},
			"body": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Text describing the contents of the tag.",
			},
			"draft": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				ForceNew:    true,
				Description: "Set to 'false' to create a published release.",
			},
			"prerelease": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Set to 'false' to identify the release as a full release.",
			},
			"generate_release_notes": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Set to 'true' to automatically generate the name and body for this release. If 'name' is specified, the specified name will be used; otherwise, a name will be automatically generated. If 'body' is specified, the body will be pre-pended to the automatically generated notes.",
			},
			"discussion_category_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "If specified, a discussion of the specified category is created and linked to the release. The value must be a category that already exists in the repository.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"release_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the release.",
			},
			"node_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The node ID of the release.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time the release was created.",
			},
			"published_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time the release was published.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL for the release.",
			},
			"html_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The HTML URL for the release.",
			},
			"assets_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL for the release assets.",
			},
			"upload_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL for the uploaded assets of release.",
			},
			"zipball_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL for the zipball of the release.",
			},
			"tarball_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL for the tarball of the release.",
			},
		},
	}
}

func resourceGithubReleaseCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	tagName := d.Get("tag_name").(string)
	targetCommitish := d.Get("target_commitish").(string)
	draft := d.Get("draft").(bool)
	prerelease := d.Get("prerelease").(bool)
	generateReleaseNotes := d.Get("generate_release_notes").(bool)

	req := &github.RepositoryRelease{
		TagName:              github.String(tagName),
		TargetCommitish:      github.String(targetCommitish),
		Draft:                github.Bool(draft),
		Prerelease:           github.Bool(prerelease),
		GenerateReleaseNotes: github.Bool(generateReleaseNotes),
	}

	if v, ok := d.GetOk("body"); ok {
		req.Body = github.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = github.String(v.(string))
	}

	if v, ok := d.GetOk("discussion_category_name"); ok {
		req.DiscussionCategoryName = github.String(v.(string))
	}

	var release *github.RepositoryRelease
	var resp *github.Response
	var err error
	if d.IsNewResource() {
		log.Printf("[DEBUG] Creating release: %s (%s/%s)",
			targetCommitish, owner, repoName)
		release, resp, err = client.Repositories.CreateRelease(ctx, owner, repoName, req)
		if resp != nil {
			log.Printf("[DEBUG] Response from creating release: %#v", *resp)
		}
	} else {
		number := d.Get("number").(int64)
		log.Printf("[DEBUG] Updating release: %d:%s (%s/%s)",
			number, targetCommitish, owner, repoName)
		release, resp, err = client.Repositories.EditRelease(ctx, owner, repoName, number, req)
		if resp != nil {
			log.Printf("[DEBUG] Response from updating release: %#v", *resp)
		}
	}

	if err != nil {
		return err
	}
	transformResponseToResourceData(d, release, repoName)
	return nil
}

func resourceGithubReleaseRead(d *schema.ResourceData, meta interface{}) error {
	repository := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	releaseID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	if releaseID == 0 {
		return fmt.Errorf("`release_id` must be present")
	}

	release, _, err := client.Repositories.GetRelease(ctx, owner, repository, releaseID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing release ID %d for repository %s from state, because it no longer exists on GitHub", releaseID, repository)
				d.SetId("")
				return nil
			}
		}
		return err
	}
	transformResponseToResourceData(d, release, repository)
	return nil
}

func resourceGithubReleaseDelete(d *schema.ResourceData, meta interface{}) error {
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	repository := d.Get("repository").(string)
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name

	releaseIDStr := d.Id()
	releaseID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(releaseIDStr, err)
	}
	if releaseID == 0 {
		return fmt.Errorf("`release_id` must be present")
	}

	_, err = client.Repositories.DeleteRelease(ctx, owner, repository, releaseID)
	if err != nil {
		return fmt.Errorf("error deleting GitHub release reference %s/%s (%s): %s",
			fmt.Sprint(releaseID), repository, owner, err)
	}
	return nil
}

func resourceGithubReleaseImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	repoName, releaseIDStr, err := parseTwoPartID(d.Id(), "repository", "release")
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	releaseID, err := strconv.ParseInt(releaseIDStr, 10, 64)
	if err != nil {
		return []*schema.ResourceData{d}, unconvertibleIdErr(releaseIDStr, err)
	}
	if releaseID == 0 {
		return []*schema.ResourceData{d}, fmt.Errorf("`release_id` must be present")
	}
	log.Printf("[DEBUG] Importing release with ID: %d, for repository: %s", releaseID, repoName)

	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()
	repository, _, err := client.Repositories.Get(ctx, owner, repoName)
	if repository == nil || err != nil {
		return []*schema.ResourceData{d}, err
	}
	if err = d.Set("repository", *repository.Name); err != nil {
		return []*schema.ResourceData{d}, err
	}

	release, _, err := client.Repositories.GetRelease(ctx, owner, *repository.Name, releaseID)
	if release == nil || err != nil {
		return []*schema.ResourceData{d}, err
	}
	d.SetId(strconv.FormatInt(release.GetID(), 10))

	return []*schema.ResourceData{d}, nil
}

func transformResponseToResourceData(d *schema.ResourceData, release *github.RepositoryRelease, repository string) {
	d.SetId(strconv.FormatInt(release.GetID(), 10))
	_ = d.Set("release_id", release.GetID())
	_ = d.Set("node_id", release.GetNodeID())
	_ = d.Set("repository", repository)
	_ = d.Set("tag_name", release.GetTagName())
	_ = d.Set("target_commitish", release.GetTargetCommitish())
	_ = d.Set("name", release.GetName())
	_ = d.Set("body", release.GetBody())
	_ = d.Set("draft", release.GetDraft())
	_ = d.Set("generate_release_notes", release.GetGenerateReleaseNotes())
	_ = d.Set("prerelease", release.GetPrerelease())
	_ = d.Set("discussion_category_name", release.GetDiscussionCategoryName())
	_ = d.Set("created_at", release.GetCreatedAt().String())
	_ = d.Set("published_at", release.GetPublishedAt().String())
	_ = d.Set("url", release.GetURL())
	_ = d.Set("html_url", release.GetHTMLURL())
	_ = d.Set("assets_url", release.GetAssetsURL())
	_ = d.Set("upload_url", release.GetUploadURL())
	_ = d.Set("zipball_url", release.GetZipballURL())
	_ = d.Set("tarball_url", release.GetTarballURL())
}
