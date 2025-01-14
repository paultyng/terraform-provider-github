package github

import (
	"context"
	"log"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubTeamOrganizationRoleAssignment() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubTeamOrganizationRoleAssignmentCreate,
		Read:   resourceGithubTeamOrganizationRoleAssignmentRead,
		Delete: resourceGithubTeamOrganizationRoleAssignmentDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				teamIdString, roleID, err := parseTwoPartID(d.Id(), "team_id", "role_id")
				if err != nil {
					return nil, err
				}

				teamSlug, err := getTeamSlug(teamIdString, meta)
				if err != nil {
					return nil, err
				}

				d.SetId(buildTwoPartID(teamSlug, roleID))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The GitHub team id or the GitHub team slug.",
				ForceNew:    true,
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The GitHub Organization Role id.",
				ForceNew:    true,
			},
		},
	}
}

func resourceGithubTeamOrganizationRoleAssignmentCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.Background()

	// The given team id could be an id or a slug
	givenTeamId := d.Get("team_id").(string)
	teamSlug, err := getTeamSlug(givenTeamId, meta)
	if err != nil {
		return err
	}

	roleIDString := d.Get("role_id").(string)
	roleID, err := strconv.ParseInt(roleIDString, 10, 32)

	if err != nil {
		return err
	}

	_, err = client.Organizations.AssignOrgRoleToTeam(ctx, orgName, teamSlug, roleID)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(teamSlug, roleIDString))
	return resourceGithubTeamOrganizationRoleAssignmentRead(d, meta)
}

func resourceGithubTeamOrganizationRoleAssignmentRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	teamIdString, roleIDString, err := parseTwoPartID(d.Id(), "team_id", "role_id")
	if err != nil {
		return err
	}

	// The given team id could be an id or a slug
	teamSlug, err := getTeamSlug(teamIdString, meta)
	if err != nil {
		return err
	}

	roleID, err := strconv.ParseInt(roleIDString, 10, 32)
	if err != nil {
		return err
	}

	// There is no api for checking a specific team role assignment, so instead we iterate over all teams assigned to the role
	// go-github pagination (https://github.com/google/go-github?tab=readme-ov-file#pagination)
	options := &github.ListOptions{
		PerPage: 100,
	}
	var foundTeam *github.Team
	for {
		teams, resp, err := client.Organizations.ListTeamsAssignedToOrgRole(ctx, orgName, roleID, options)
		if err != nil {
			return err
		}

		for _, team := range teams {
			if team.GetSlug() == teamSlug {
				foundTeam = team
				break
			}

		}

		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	if foundTeam == nil {
		log.Printf("[WARN] Removing team organization role association %s from state because it no longer exists in GitHub", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceGithubTeamOrganizationRoleAssignmentDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.Background()

	teamIdString, roleIDString, err := parseTwoPartID(d.Id(), "team_id", "role_id")
	if err != nil {
		return err
	}

	// The given team id could be an id or a slug
	teamSlug, err := getTeamSlug(teamIdString, meta)
	if err != nil {
		return err
	}

	roleID, err := strconv.ParseInt(roleIDString, 10, 32)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	_, err = client.Organizations.RemoveOrgRoleFromTeam(ctx, orgName, teamSlug, roleID)
	if err != nil {
		return err
	}

	return nil
}
