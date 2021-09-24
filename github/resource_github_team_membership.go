package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubTeamMembership() *schema.Resource {

	return &schema.Resource{
		Create: resourceGithubTeamMembershipCreateOrUpdate,
		Read:   resourceGithubTeamMembershipRead,
		Update: resourceGithubTeamMembershipCreateOrUpdate,
		Delete: resourceGithubTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTeamIDFunc,
			},
			"username": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: caseInsensitive(),
			},
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "member",
				ValidateFunc: validateValueFunc([]string{"member", "maintainer"}),
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubTeamMembershipCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id

	teamIdString := d.Get("team_id").(string)
	teamId, err := strconv.ParseInt(teamIdString, 10, 64)
	if err != nil {
		return unconvertibleIdErr(teamIdString, err)
	}
	ctx := context.Background()

	username := d.Get("username").(string)
	role := d.Get("role").(string)

	log.Printf("[DEBUG] Creating team membership: %s/%s (%s)", teamIdString, username, role)
	_, _, err = client.Teams.AddTeamMembershipByID(ctx,
		orgId,
		teamId,
		username,
		&github.TeamAddTeamMembershipOptions{
			Role: role,
		},
	)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(teamIdString, username))

	return resourceGithubTeamMembershipRead(d, meta)
}

func resourceGithubTeamMembershipRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id
	teamIdString, username, err := parseTwoPartID(d.Id(), "team_id", "username")
	if err != nil {
		return err
	}

	teamId, err := strconv.ParseInt(teamIdString, 10, 64)
	if err != nil {
		return unconvertibleIdErr(teamIdString, err)
	}

	// We intentionally set these early to allow reconciliation
	// from an upstream bug which emptied team_id in state
	// See https://github.com/integrations/terraform-provider-github/issues/323
	d.Set("team_id", teamIdString)
	d.Set("username", username)

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	log.Printf("[DEBUG] Reading team membership: %s/%s", teamIdString, username)
	membership, resp, err := client.Teams.GetTeamMembershipByID(ctx,
		orgId, teamId, username)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing team membership %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("role", membership.GetRole())

	return nil
}

func resourceGithubTeamMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id
	teamIdString := d.Get("team_id").(string)
	teamId, err := strconv.ParseInt(teamIdString, 10, 64)
	if err != nil {
		return unconvertibleIdErr(teamIdString, err)
	}
	username := d.Get("username").(string)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Deleting team membership: %s/%s", teamIdString, username)
	_, err = client.Teams.RemoveTeamMembershipByID(ctx, orgId, teamId, username)

	return err
}
