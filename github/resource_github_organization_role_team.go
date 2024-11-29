package github

import (
	"context"
	"log"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubOrganizationRoleTeam() *schema.Resource {
	return &schema.Resource{
		Description: "Manage an association between an organization role and a team.",

		Create: resourceGithubOrganizationRoleTeamCreate,
		Read:   resourceGithubOrganizationRoleTeamRead,
		Delete: resourceGithubOrganizationRoleTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "The ID of the organization role.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"team_slug": {
				Description: "The slug of the team name.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceGithubOrganizationRoleTeamCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleId := int64(d.Get("role_id").(int))
	teamSlug := d.Get("team_slug").(string)

	_, err = client.Organizations.AssignOrgRoleToTeam(ctx, orgName, teamSlug, roleId)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(strconv.FormatInt(roleId, 10), teamSlug))

	return resourceGithubOrganizationRoleTeamRead(d, meta)
}

func resourceGithubOrganizationRoleTeamRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleIdString, teamSlug, err := parseTwoPartID(d.Id(), "role_id", "team_slug")
	if err != nil {
		return err
	}
	roleId, err := strconv.ParseInt(roleIdString, 10, 64)
	if err != nil {
		return err
	}

	opts := &github.ListOptions{
		PerPage: maxPerPage,
	}

	var team *github.Team
	for {
		teams, resp, err := client.Organizations.ListTeamsAssignedToOrgRole(ctx, orgName, roleId, opts)
		if err != nil {
			return err
		}

		for _, t := range teams {
			if t.GetSlug() == teamSlug {
				team = t
				break
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if team == nil {
		log.Printf("[INFO] Removing organization role team (%d:%s) from state because it no longer exists in GitHub", roleId, teamSlug)
		d.SetId("")
		return nil
	}

	if err = d.Set("role_id", roleId); err != nil {
		return err
	}

	if err = d.Set("team_slug", teamSlug); err != nil {
		return err
	}

	return nil
}

func resourceGithubOrganizationRoleTeamDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleId := int64(d.Get("role_id").(int))
	teamSlug := d.Get("team_slug").(string)

	_, err = client.Organizations.RemoveOrgRoleFromTeam(ctx, orgName, teamSlug, roleId)
	return err
}
