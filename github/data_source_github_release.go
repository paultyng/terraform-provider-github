package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{
					"latest",
					"id",
					"tag",
				}, false), "retrieve_by"),
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
			"assets_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asserts_url": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "use assets_url instead",
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
			"assets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"node_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"content_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"browser_download_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
		release, _, err = client.Repositories.GetLatestRelease(ctx, owner, repository)
	case "id":
		releaseID := int64(d.Get("release_id").(int))
		if releaseID == 0 {
			return fmt.Errorf("`release_id` must be set when `retrieve_by` = `id`")
		}

		release, _, err = client.Repositories.GetRelease(ctx, owner, repository, releaseID)
	case "tag":
		tag := d.Get("release_tag").(string)
		if tag == "" {
			return fmt.Errorf("`release_tag` must be set when `retrieve_by` = `tag`")
		}

		release, _, err = client.Repositories.GetReleaseByTag(ctx, owner, repository, tag)
	default:
		return fmt.Errorf("one of: `latest`, `id`, `tag` must be set for `retrieve_by`")
	}

	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(release.GetID(), 10))
	err = d.Set("release_tag", release.GetTagName())
	if err != nil {
		return err
	}
	err = d.Set("target_commitish", release.GetTargetCommitish())
	if err != nil {
		return err
	}
	err = d.Set("name", release.GetName())
	if err != nil {
		return err
	}
	err = d.Set("body", release.GetBody())
	if err != nil {
		return err
	}
	err = d.Set("draft", release.GetDraft())
	if err != nil {
		return err
	}
	err = d.Set("prerelease", release.GetPrerelease())
	if err != nil {
		return err
	}
	err = d.Set("created_at", release.GetCreatedAt().String())
	if err != nil {
		return err
	}
	err = d.Set("published_at", release.GetPublishedAt().String())
	if err != nil {
		return err
	}
	err = d.Set("url", release.GetURL())
	if err != nil {
		return err
	}
	err = d.Set("html_url", release.GetHTMLURL())
	if err != nil {
		return err
	}
	err = d.Set("assets_url", release.GetAssetsURL())
	if err != nil {
		return err
	}
	err = d.Set("asserts_url", release.GetAssetsURL()) // Deprecated, original version of assets_url
	if err != nil {
		return err
	}
	err = d.Set("upload_url", release.GetUploadURL())
	if err != nil {
		return err
	}
	err = d.Set("zipball_url", release.GetZipballURL())
	if err != nil {
		return err
	}
	err = d.Set("tarball_url", release.GetTarballURL())
	if err != nil {
		return err
	}

	assets := make([]interface{}, 0, len(release.Assets))
	for _, releaseAsset := range release.Assets {
		if releaseAsset == nil {
			continue
		}

		assets = append(assets, map[string]interface{}{
			"id":                   releaseAsset.GetID(),
			"url":                  releaseAsset.GetURL(),
			"node_id":              releaseAsset.GetNodeID(),
			"name":                 releaseAsset.GetName(),
			"label":                releaseAsset.GetLabel(),
			"content_type":         releaseAsset.GetContentType(),
			"size":                 releaseAsset.GetSize(),
			"created_at":           releaseAsset.GetCreatedAt().String(),
			"updated_at":           releaseAsset.GetUpdatedAt().String(),
			"browser_download_url": releaseAsset.GetBrowserDownloadURL(),
		})
	}

	err = d.Set("assets", assets)
	if err != nil {
		return err
	}

	return nil
}
