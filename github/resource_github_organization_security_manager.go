package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubOrganizationSecurityManager() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubOrganizationSecurityManagerCreate,
		Read:   resourceGithubOrganizationSecurityManagerRead,
		Update: resourceGithubOrganizationSecurityManagerUpdate,
		Delete: resourceGithubOrganizationSecurityManagerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"team_slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the team to manage.",
			},
		},
	}
}

func resourceGithubOrganizationSecurityManagerCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	orgName := meta.(*Owner).name
	teamSlug := d.Get("team_slug").(string)

	client := meta.(*Owner).v3client
	ctx := context.Background()

	team, _, err := client.Teams.GetTeamBySlug(ctx, orgName, teamSlug)
	if err != nil {
		log.Printf("[INFO] Team %s/%s was not found in GitHub", orgName, teamSlug)
		return err
	}

	_, err = client.Organizations.AddSecurityManagerTeam(ctx, orgName, teamSlug)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusConflict {
				log.Printf("[WARN] Organization %s has reached the maximum number of security manager teams", orgName)
				return nil
			}
		}
		return err
	}

	d.SetId(strconv.FormatInt(team.GetID(), 10))

	return resourceGithubOrganizationSecurityManagerRead(d, meta)
}

func resourceGithubOrganizationSecurityManagerRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	orgName := meta.(*Owner).name
	teamId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	// There is no endpoint for getting a single security manager team, so get the list and filter.
	// There is a maximum number of security manager teams (currently 10), so this should be fine.
	teams, _, err := client.Organizations.ListSecurityManagerTeams(ctx, orgName)
	if err != nil {
		return err
	}

	var team *github.Team
	for _, t := range teams {
		if t.GetID() == teamId {
			team = t
			break
		}
	}

	if team == nil {
		log.Printf("[WARN] Removing organization security manager team %s from state because it no longer exists in GitHub", d.Id())
		d.SetId("")
		return nil
	}

	if err = d.Set("team_slug", team.GetSlug()); err != nil {
		return err
	}

	return nil
}

func resourceGithubOrganizationSecurityManagerUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	orgId := meta.(*Owner).id
	orgName := meta.(*Owner).name
	teamId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	team, _, err := client.Teams.GetTeamByID(ctx, orgId, teamId)
	if err != nil {
		return err
	}

	// Adding the same team is a no-op.
	_, err = client.Organizations.AddSecurityManagerTeam(ctx, orgName, team.GetSlug())
	if err != nil {
		return err
	}

	return resourceGithubOrganizationSecurityManagerRead(d, meta)
}

func resourceGithubOrganizationSecurityManagerDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	orgName := meta.(*Owner).name
	teamSlug := d.Get("team_slug").(string)

	client := meta.(*Owner).v3client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Organizations.RemoveSecurityManagerTeam(ctx, orgName, teamSlug)
	return err
}
