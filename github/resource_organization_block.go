package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceOrganizationBlock() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationBlockCreate,
		Read:   resourceOrganizationBlockRead,
		Delete: resourceOrganizationBlockDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOrganizationBlockCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.Background()
	username := d.Get("username").(string)

	log.Printf("[DEBUG] Creating organization block: %s (%s)", username, orgName)
	_, err = client.Organizations.BlockUser(ctx, orgName, username)
	if err != nil {
		return err
	}
	d.SetId(username)

	return resourceOrganizationBlockRead(d, meta)
}

func resourceOrganizationBlockRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	username := d.Id()

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	log.Printf("[DEBUG] Reading organization block: %s (%s)", d.Id(), orgName)
	blocked, resp, err := client.Organizations.IsBlocked(ctx, orgName, username)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			// not sure if this will ever be hit, I imagine just returns false?
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing organization block %s/%s from state because it no longer exists in GitHub",
					orgName, d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if !blocked {
		d.SetId("")
		return nil
	}

	d.Set("username", username)
	d.Set("etag", resp.Header.Get("ETag"))

	return nil
}

func resourceOrganizationBlockDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	username := d.Id()
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting organization block: %s (%s)", d.Id(), orgName)
	_, err := client.Organizations.UnblockUser(ctx, orgName, username)
	return err
}
