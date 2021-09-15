package github

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubUserSshKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubUserSshKeyCreate,
		Read:   resourceGithubUserSshKeyRead,
		Delete: resourceGithubUserSshKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"title": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, oldV, newV string, d *schema.ResourceData) bool {
					newTrimmed := strings.TrimSpace(newV)
					return oldV == newTrimmed
				},
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
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

	log.Printf("[DEBUG] Creating user SSH key: %s", title)
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

	log.Printf("[DEBUG] Reading user SSH key: %s", d.Id())
	key, resp, err := client.Users.GetKey(ctx, id)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing user SSH key %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("title", key.GetTitle())
	d.Set("key", key.GetKey())
	d.Set("url", key.GetURL())

	return nil
}

func resourceGithubUserSshKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting user SSH key: %s", d.Id())
	_, err = client.Users.DeleteKey(ctx, id)

	return err
}
