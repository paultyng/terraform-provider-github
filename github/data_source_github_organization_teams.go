package github

import (
	"context"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/shurcooL/githubv4"
)

func dataSourceGithubOrganizationTeams() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubOrganizationTeamsRead,

		Schema: map[string]*schema.Schema{
			"root_teams_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"summary_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"results_per_page": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          100,
				ValidateDiagFunc: toDiagFunc(validation.IntBetween(0, 100), "results_per_page"),
			},
			"teams": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"node_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"privacy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"members": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"repositories": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"parent": {
							Deprecated: "Use parent_team_id and parent_team_slug instead.",
							Type:       schema.TypeMap,
							Computed:   true,
							Elem:       &schema.Schema{Type: schema.TypeString},
						},
						"parent_team_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parent_team_slug": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubOrganizationTeamsRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	clientv3 := meta.(*Owner).v3client
	client := meta.(*Owner).v4client
	orgName := meta.(*Owner).name
	rootTeamsOnly := d.Get("root_teams_only").(bool)
	summaryOnly := d.Get("summary_only").(bool)
	resultsPerPage := d.Get("results_per_page").(int)

	var query TeamsQuery

	variables := map[string]interface{}{
		"first":         githubv4.Int(resultsPerPage),
		"login":         githubv4.String(orgName),
		"cursor":        (*githubv4.String)(nil),
		"rootTeamsOnly": githubv4.Boolean(rootTeamsOnly),
		"summaryOnly":   githubv4.Boolean(summaryOnly),
	}

	var teams []interface{}
	for {
		err = client.Query(meta.(*Owner).StopContext, &query, variables)
		if err != nil {
			return err
		}

		additionalTeams, err := flattenGitHubTeams(clientv3, meta.(*Owner).StopContext, orgName, query)
		if err != nil {
			return err
		}
		teams = append(teams, additionalTeams...)

		if !query.Organization.Teams.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(query.Organization.Teams.PageInfo.EndCursor)
	}

	d.SetId(string(query.Organization.ID))
	err = d.Set("teams", teams)
	if err != nil {
		return err
	}

	return nil
}

func flattenGitHubTeams(client *github.Client, ctx context.Context, org string, tq TeamsQuery) ([]interface{}, error) {
	teams := tq.Organization.Teams.Nodes

	if len(teams) == 0 {
		return make([]interface{}, 0), nil
	}

	flatTeams := make([]interface{}, len(teams))

	for i, team := range teams {
		t := make(map[string]interface{})

		t["id"] = team.DatabaseID
		t["node_id"] = team.ID
		t["slug"] = team.Slug
		t["name"] = team.Name
		t["description"] = team.Description
		t["privacy"] = team.Privacy
		members := team.Members.Nodes
		flatMembers := make([]string, len(members))

		for i, member := range members {
			flatMembers[i] = string(member.Login)
		}

		t["members"] = flatMembers

		var parentTeamId string
		if len(team.Parent.Slug) != 0 {
			parentTeam, _, err := client.Teams.GetTeamBySlug(ctx, org, string(team.Parent.Slug))
			if err != nil {
				return nil, err
			}
			parentTeamId = strconv.FormatInt(parentTeam.GetID(), 10)
		}

		t["parent_team_id"] = parentTeamId
		t["parent_team_slug"] = team.Parent.Slug

		parentTeam := make(map[string]interface{})
		parentTeam["id"] = team.Parent.ID
		parentTeam["slug"] = team.Parent.Slug
		parentTeam["name"] = team.Parent.Name
		t["parent"] = parentTeam

		repositories := team.Repositories.Nodes

		flatRepositories := make([]string, len(repositories))

		for i, repository := range repositories {
			flatRepositories[i] = string(repository.Name)
		}

		t["repositories"] = flatRepositories

		flatTeams[i] = t
	}

	return flatTeams, nil
}
