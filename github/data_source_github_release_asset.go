package github

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubReleaseAsset() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubReleaseAssetRead,

		Schema: map[string]*schema.Schema{
			"asset_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_id": {
				Type:     schema.TypeString,
				Computed: true,
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
	}
}

func dataSourceGithubReleaseAssetRead(d *schema.ResourceData, meta interface{}) error {
	repository := d.Get("repository").(string)
	owner := d.Get("owner").(string)

	client := meta.(*Owner).v3client
	ctx := context.Background()

	var err error
	var asset *github.ReleaseAsset

	assetID := int64(d.Get("asset_id").(int))
	asset, _, err = client.Repositories.GetReleaseAsset(ctx, owner, repository, assetID)
	if err != nil {
		return err
	}

	var respBody io.ReadCloser
	clientCopy := client.Client()
	respBody, _, err = client.Repositories.DownloadReleaseAsset(ctx, owner, repository, assetID, clientCopy)
	if err != nil {
		return err
	}
	defer respBody.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, respBody)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(asset.GetID(), 10))
	err = d.Set("body", buf.String())
	if err != nil {
		return err
	}
	err = d.Set("url", asset.URL)
	if err != nil {
		return err
	}
	err = d.Set("node_id", asset.NodeID)
	if err != nil {
		return err
	}
	err = d.Set("name", asset.Name)
	if err != nil {
		return err
	}
	err = d.Set("label", asset.Label)
	if err != nil {
		return err
	}
	err = d.Set("content_type", asset.ContentType)
	if err != nil {
		return err
	}
	err = d.Set("size", asset.Size)
	if err != nil {
		return err
	}
	err = d.Set("created_at", asset.CreatedAt.String())
	if err != nil {
		return err
	}
	err = d.Set("created_at", asset.UpdatedAt.String())
	if err != nil {
		return err
	}
	err = d.Set("browser_download_url", asset.BrowserDownloadURL)
	if err != nil {
		return err
	}

	return nil
}
