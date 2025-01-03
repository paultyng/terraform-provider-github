package github

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/githubv4"
)

func resourceGithubTeamSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubTeamSettingsCreate,
		Read:   resourceGithubTeamSettingsRead,
		Update: resourceGithubTeamSettingsUpdate,
		Delete: resourceGithubTeamSettingsDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubTeamSettingsImport,
		},
		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub team id or the GitHub team slug.",
			},
			"team_slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The slug of the Team within the Organization.",
			},
			"team_uid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique ID of the Team on GitHub. Corresponds to the ID of the 'github_team_settings' resource.",
			},
			"review_request_delegation": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "The settings for delegating code reviews to individuals on behalf of the team. If this block is present, even without any fields, then review request delegation will be enabled for the team.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"algorithm": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The algorithm to use when assigning pull requests to team members. Supported values are 'ROUND_ROBIN' and 'LOAD_BALANCE'.",
							Default:     "ROUND_ROBIN",
							ValidateDiagFunc: toDiagFunc(func(v interface{}, key string) (we []string, errs []error) {
								algorithm, ok := v.(string)
								if !ok {
									return nil, []error{fmt.Errorf("expected type of %s to be string", key)}
								}

								if !(algorithm == "ROUND_ROBIN" || algorithm == "LOAD_BALANCE") {
									errs = append(errs, errors.New("review request delegation algorithm must be one of [\"ROUND_ROBIN\", \"LOAD_BALANCE\"]"))
								}

								return we, errs
							}, "algorithm"),
						},
						"member_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							RequiredWith: []string{"review_request_delegation"},
							Description:  "The number of team members to assign to a pull request.",
							ValidateDiagFunc: toDiagFunc(func(v interface{}, key string) (we []string, errs []error) {
								count, ok := v.(int)
								if !ok {
									return nil, []error{fmt.Errorf("expected type of %s to be an integer", key)}
								}
								if count <= 0 {
									errs = append(errs, errors.New("review request delegation reviewer count must be a positive number"))
								}
								return we, errs
							}, "member_count"),
						},
						"notify": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "whether to notify the entire team when at least one member is also assigned to the pull request.",
						},
						// TODO: The underlying graphql call is using memberIds. Might be better UX to use usernames, or add both as an option...
						"excluded_team_member_node_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "A list of users to exclude from the PR review process",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceGithubTeamSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	// Given a string that is either a team id or team slug, return the
	// get the basic details of the team including node_id and slug
	ctx := context.Background()

	teamIDString, _ := d.Get("team_id").(string)

	nodeId, slug, err := resolveTeamIDs(teamIDString, meta.(*Owner), ctx)
	if err != nil {
		return err
	}
	d.SetId(nodeId)
	if err = d.Set("team_slug", slug); err != nil {
		return err
	}
	if err = d.Set("team_uid", nodeId); err != nil {
		return err
	}
	return resourceGithubTeamSettingsUpdate(d, meta)

}

func resourceGithubTeamSettingsRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	graphql := meta.(*Owner).v4client
	orgName := meta.(*Owner).name

	teamSlug := d.Get("team_slug").(string)

	var query = queryTeamSettings{}
	variables := map[string]interface{}{
		"slug":  githubv4.String(teamSlug),
		"login": githubv4.String(orgName),
	}

	e := graphql.Query(meta.(*Owner).StopContext, &query, variables)
	if e != nil {
		return e
	}

	if query.Organization.Team.ReviewRequestDelegation {
		reviewRequestDelegation := make(map[string]interface{})
		reviewRequestDelegation["algorithm"] = query.Organization.Team.ReviewRequestDelegationAlgorithm
		reviewRequestDelegation["member_count"] = query.Organization.Team.ReviewRequestDelegationCount
		reviewRequestDelegation["notify"] = query.Organization.Team.ReviewRequestDelegationNotifyAll
		if err = d.Set("review_request_delegation", []interface{}{reviewRequestDelegation}); err != nil {
			return err
		}
	} else {
		if err = d.Set("review_request_delegation", []interface{}{}); err != nil {
			return err
		}
	}
	// TODO: The exlusion list is not available via the GraphQL call yet (https://docs.github.com/en/enterprise-cloud@latest/graphql/overview/schema-previews#team-review-assignments-preview). Leaving that for future implementation

	return nil

}

func resourceGithubTeamSettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("review_request_delegation") || d.IsNewResource() {

		ctx := context.WithValue(context.Background(), ctxId, d.Id())
		graphql := meta.(*Owner).v4client
		if setting := d.Get("review_request_delegation").([]interface{}); len(setting) == 0 {
			var mutation struct {
				UpdateTeamReviewAssignment struct {
					ClientMutationId githubv4.ID `graphql:"clientMutationId"`
				} `graphql:"updateTeamReviewAssignment(input:$input)"`
			}

			return graphql.Mutate(ctx, &mutation, defaultTeamReviewAssignmentSettings(d.Id()), nil)
		} else {
			settings := d.Get("review_request_delegation").([]interface{})[0].(map[string]interface{})

			var mutation struct {
				UpdateTeamReviewAssignment struct {
					ClientMutationId githubv4.ID `graphql:"clientMutationId"`
				} `graphql:"updateTeamReviewAssignment(input:$input)"`
			}

			exclusionList := make([]string, 0)
			for _, v := range settings["excluded_team_member_node_ids"].(*schema.Set).List() {
				if v != nil {
					exclusionList = append(exclusionList, v.(string))
				}
			}

			return graphql.Mutate(ctx, &mutation, UpdateTeamReviewAssignmentInput{
				TeamID:                             d.Id(),
				ReviewRequestDelegation:            true,
				ReviewRequestDelegationAlgorithm:   settings["algorithm"].(string),
				ReviewRequestDelegationCount:       settings["member_count"].(int),
				ReviewRequestDelegationNotifyAll:   settings["notify"].(bool),
				ReviewRequestexcludedTeamMemberIds: exclusionList,
			}, nil)
		}
	}

	return resourceGithubTeamSettingsRead(d, meta)

}

func resourceGithubTeamSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	graphql := meta.(*Owner).v4client

	var mutation struct {
		UpdateTeamReviewAssignment struct {
			ClientMutationId githubv4.ID `graphql:"clientMutationId"`
		} `graphql:"updateTeamReviewAssignment(input:$input)"`
	}

	return graphql.Mutate(ctx, &mutation, defaultTeamReviewAssignmentSettings(d.Id()), nil)

}

func resourceGithubTeamSettingsImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	nodeId, slug, err := resolveTeamIDs(d.Id(), meta.(*Owner), context.Background())
	if err != nil {
		return nil, err
	}
	if err = d.Set("team_id", d.Id()); err != nil {
		return nil, err
	}
	d.SetId(nodeId)
	if err = d.Set("team_slug", slug); err != nil {
		return nil, err
	}
	if err = d.Set("team_uid", nodeId); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, resourceGithubTeamSettingsRead(d, meta)
}

func resolveTeamIDs(idOrSlug string, meta *Owner, ctx context.Context) (nodeId string, slug string, err error) {
	client := meta.v3client
	orgName := meta.name
	orgId := meta.id

	teamId, parseIntErr := strconv.ParseInt(idOrSlug, 10, 64)
	if parseIntErr != nil {
		// The given id not an integer, assume it is a team slug
		team, _, slugErr := client.Teams.GetTeamBySlug(ctx, orgName, idOrSlug)
		if slugErr != nil {
			return "", "", errors.New(parseIntErr.Error() + slugErr.Error())
		}
		return team.GetNodeID(), team.GetSlug(), nil
	} else {
		// The given id is an integer, assume it is a team id
		team, _, teamIdErr := client.Teams.GetTeamByID(ctx, orgId, teamId)
		if teamIdErr != nil {
			// There isn't a team with the given ID, assume it is a teamslug
			team, _, slugErr := client.Teams.GetTeamBySlug(ctx, orgName, idOrSlug)
			if slugErr != nil {
				return "", "", errors.New(teamIdErr.Error() + slugErr.Error())
			}

			return team.GetNodeID(), team.GetSlug(), nil
		}

		return team.GetNodeID(), team.GetSlug(), nil
	}
}

type UpdateTeamReviewAssignmentInput struct {
	ClientMutationID                   string   `json:"clientMutationId,omitempty"`
	TeamID                             string   `graphql:"id" json:"id"`
	ReviewRequestDelegation            bool     `graphql:"enabled" json:"enabled"`
	ReviewRequestDelegationAlgorithm   string   `graphql:"algorithm" json:"algorithm"`
	ReviewRequestDelegationCount       int      `graphql:"teamMemberCount" json:"teamMemberCount"`
	ReviewRequestDelegationNotifyAll   bool     `graphql:"notifyTeam" json:"notifyTeam"`
	ReviewRequestexcludedTeamMemberIds []string `graphql:"excludedTeamMemberIds" json:"excludedTeamMemberIds"`
}

func defaultTeamReviewAssignmentSettings(id string) UpdateTeamReviewAssignmentInput {
	return UpdateTeamReviewAssignmentInput{
		TeamID:                           id,
		ReviewRequestDelegation:          false,
		ReviewRequestDelegationAlgorithm: "ROUND_ROBIN",
		ReviewRequestDelegationCount:     1,
		ReviewRequestDelegationNotifyAll: true,
	}
}

type queryTeamSettings struct {
	Organization struct {
		Team struct {
			Name                             string `graphql:"name"`
			Slug                             string `graphql:"slug"`
			ID                               string `graphql:"id"`
			ReviewRequestDelegation          bool   `graphql:"reviewRequestDelegationEnabled"`
			ReviewRequestDelegationAlgorithm string `graphql:"reviewRequestDelegationAlgorithm"`
			ReviewRequestDelegationCount     int    `graphql:"reviewRequestDelegationMemberCount"`
			ReviewRequestDelegationNotifyAll bool   `graphql:"reviewRequestDelegationNotifyTeam"`
		} `graphql:"team(slug:$slug)"`
	} `graphql:"organization(login:$login)"`
}
