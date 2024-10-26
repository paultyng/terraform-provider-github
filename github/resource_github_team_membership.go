package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubTeamMembership() *schema.Resource {

	return &schema.Resource{
		Create: resourceGithubTeamMembershipCreateOrUpdate,
		Read:   resourceGithubTeamMembershipRead,
		Update: resourceGithubTeamMembershipCreateOrUpdate,
		Delete: resourceGithubTeamMembershipDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				teamIdString, username, err := parseTwoPartID(d.Id(), "team_id", "username")
				if err != nil {
					return nil, err
				}

				teamId, err := getTeamID(teamIdString, meta)
				if err != nil {
					return nil, err
				}

				d.SetId(buildTwoPartID(strconv.FormatInt(teamId, 10), username))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub team id or the GitHub team slug.",
			},
			"username": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: caseInsensitive(),
				Description:      "The user to add to the team.",
			},
			"role": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "member",
				Description:      "The role of the user within the team. Must be one of 'member' or 'maintainer'.",
				ValidateDiagFunc: validateValueFunc([]string{"member", "maintainer"}),
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
	teamId, err := getTeamID(teamIdString, meta)
	if err != nil {
		return err
	}
	ctx := context.Background()

	username := d.Get("username").(string)
	role := d.Get("role").(string)

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

	teamId, err := getTeamID(teamIdString, meta)
	if err != nil {
		return err
	}

	// We intentionally set these early to allow reconciliation
	// from an upstream bug which emptied team_id in state
	// See https://github.com/integrations/terraform-provider-github/issues/323
	if err = d.Set("team_id", teamIdString); err != nil {
		return err
	}
	if err = d.Set("username", username); err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	membership, resp, err := client.Teams.GetTeamMembershipByID(ctx,
		orgId, teamId, username)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing team membership %s from state because it no longer exists in GitHub",
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
	if err = d.Set("role", membership.GetRole()); err != nil {
		return err
	}

	return nil
}

func resourceGithubTeamMembershipDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id
	teamIdString := d.Get("team_id").(string)
	teamId, err := getTeamID(teamIdString, meta)
	if err != nil {
		return err
	}
	username := d.Get("username").(string)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Teams.RemoveTeamMembershipByID(ctx, orgId, teamId, username)

	return err
}
