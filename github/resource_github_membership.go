package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubMembershipCreateOrUpdate,
		Read:   resourceGithubMembershipRead,
		Update: resourceGithubMembershipCreateOrUpdate,
		Delete: resourceGithubMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: caseInsensitive(),
			},
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateValueFunc([]string{"member", "admin"}),
				Default:      "member",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubMembershipCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	username := d.Get("username").(string)
	roleName := d.Get("role").(string)
	ctx := context.Background()
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxId, d.Id())
	}

	log.Printf("[DEBUG] Creating membership: %s/%s", orgName, username)
	_, _, err = client.Organizations.EditOrgMembership(ctx,
		username,
		orgName,
		&github.Membership{
			Role: github.String(roleName),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(orgName, username))

	return resourceGithubMembershipRead(d, meta)
}

func resourceGithubMembershipRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
	_, username, err := parseTwoPartID(d.Id(), "organization", "username")
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	log.Printf("[DEBUG] Reading membership: %s", d.Id())
	membership, resp, err := client.Organizations.GetOrgMembership(ctx,
		username, orgName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing membership %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("username", username)
	d.Set("role", membership.GetRole())

	return nil
}

func resourceGithubMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting membership: %s", d.Id())
	_, err = client.Organizations.RemoveOrgMembership(ctx,
		d.Get("username").(string), orgName)

	return err
}
