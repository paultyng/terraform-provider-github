package github

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubUserSshKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubUserSshKeyCreate,
		Read:   resourceGithubUserSshKeyRead,
		Delete: resourceGithubUserSshKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A descriptive name for the new key.",
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The public SSH key to add to your GitHub account.",
				DiffSuppressFunc: func(k, oldV, newV string, d *schema.ResourceData) bool {
					newTrimmed := strings.TrimSpace(newV)
					return oldV == newTrimmed
				},
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the SSH key.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubUserSshKeyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	title := d.Get("title").(string)
	key := d.Get("key").(string)
	ctx := context.Background()

	userKey, _, err := client.Users.CreateKey(ctx, &github.Key{
		Title: github.String(title),
		Key:   github.String(key),
	})
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(*userKey.ID, 10))

	return resourceGithubUserSshKeyRead(d, meta)
}

func resourceGithubUserSshKeyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	key, resp, err := client.Users.GetKey(ctx, id)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing user SSH key %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("title", key.GetTitle()); err != nil {
		return err
	}
	if err = d.Set("key", key.GetKey()); err != nil {
		return err
	}
	if err = d.Set("url", key.GetURL()); err != nil {
		return err
	}

	return nil
}

func resourceGithubUserSshKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Users.DeleteKey(ctx, id)
	return err
}
