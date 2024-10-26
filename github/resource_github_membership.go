package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubMembershipCreateOrUpdate,
		Read:   resourceGithubMembershipRead,
		Update: resourceGithubMembershipCreateOrUpdate,
		Delete: resourceGithubMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: caseInsensitive(),
				Description:      "The user to add to the organization.",
			},
			"role": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateValueFunc([]string{"member", "admin"}),
				Default:          "member",
				Description:      "The role of the user within the organization. Must be one of 'member' or 'admin'.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"downgrade_on_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Instead of removing the member from the org, you can choose to downgrade their membership to 'member' when this resource is destroyed. This is useful when wanting to downgrade admins while keeping them in the organization",
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

	membership, resp, err := client.Organizations.GetOrgMembership(ctx,
		username, orgName)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing membership %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("username", username); err != nil {
		return err
	}
	if err = d.Set("role", membership.GetRole()); err != nil {
		return err
	}

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

	username := d.Get("username").(string)
	downgradeOnDestroy := d.Get("downgrade_on_destroy").(bool)
	downgradeTo := "member"

	if downgradeOnDestroy {
		log.Printf("[INFO] Downgrading '%s' membership for '%s' to '%s'", orgName, username, downgradeTo)

		// Check to make sure this member still has access to the organization before downgrading.
		// If we don't do this, the member would just be re-added to the organization.
		var membership *github.Membership
		membership, _, err = client.Organizations.GetOrgMembership(ctx, username, orgName)
		if err != nil {
			if ghErr, ok := err.(*github.ErrorResponse); ok {
				if ghErr.Response.StatusCode == http.StatusNotFound {
					log.Printf("[INFO] Not downgrading '%s' membership for '%s' because they are not a member of the org anymore", orgName, username)
					return nil
				}
			}

			return err
		}

		if *membership.Role == downgradeTo {
			log.Printf("[INFO] Not downgrading '%s' membership for '%s' because they are already '%s'", orgName, username, downgradeTo)
			return nil
		}

		_, _, err = client.Organizations.EditOrgMembership(ctx, username, orgName, &github.Membership{
			Role: github.String(downgradeTo),
		})
	} else {
		log.Printf("[INFO] Revoking '%s' membership for '%s'", orgName, username)
		_, err = client.Organizations.RemoveOrgMembership(ctx, username, orgName)
	}

	return err
}
