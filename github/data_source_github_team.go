package github

import (
	"context"
	"strconv"

	"github.com/google/go-github/v65/github"

	"github.com/shurcooL/githubv4"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceGithubTeam() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubTeamRead,

		Schema: map[string]*schema.Schema{
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
			"permission": {
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
			"repositories_detailed": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repo_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"role_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"node_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"membership_type": {
				Type:             schema.TypeString,
				Default:          "all",
				Optional:         true,
				ValidateDiagFunc: toDiagFunc(validation.StringInSlice([]string{"all", "immediate"}, false), "membership_type"),
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
		},
	}
}

func dataSourceGithubTeamRead(d *schema.ResourceData, meta interface{}) error {
	slug := d.Get("slug").(string)

	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id
	ctx := context.Background()
	summaryOnly := d.Get("summary_only").(bool)
	resultsPerPage := d.Get("results_per_page").(int)

	team, _, err := client.Teams.GetTeamBySlug(ctx, meta.(*Owner).name, slug)
	if err != nil {
		return err
	}

	var members []string
	var repositories []string
	var repositories_detailed []interface{}

	if !summaryOnly {
		options := github.TeamListTeamMembersOptions{
			ListOptions: github.ListOptions{
				PerPage: resultsPerPage,
			},
		}

		if d.Get("membership_type").(string) == "all" {
			for {
				member, resp, err := client.Teams.ListTeamMembersByID(ctx, orgId, team.GetID(), &options)
				if err != nil {
					return err
				}

				for _, v := range member {
					members = append(members, v.GetLogin())
				}

				if resp.NextPage == 0 {
					break
				}
				options.Page = resp.NextPage
			}
		} else {
			type member struct {
				Login string
			}
			var query struct {
				Organization struct {
					Team struct {
						Members struct {
							Nodes    []member
							PageInfo struct {
								EndCursor   githubv4.String
								HasNextPage bool
							}
						} `graphql:"members(first:100,after:$memberCursor,membership:IMMEDIATE)"`
					} `graphql:"team(slug:$slug)"`
				} `graphql:"organization(login:$owner)"`
			}
			variables := map[string]interface{}{
				"owner":        githubv4.String(meta.(*Owner).name),
				"slug":         githubv4.String(slug),
				"memberCursor": (*githubv4.String)(nil),
			}
			client := meta.(*Owner).v4client
			for {
				nameErr := client.Query(ctx, &query, variables)
				if nameErr != nil {
					return nameErr
				}
				for _, v := range query.Organization.Team.Members.Nodes {
					members = append(members, v.Login)
				}
				if query.Organization.Team.Members.PageInfo.HasNextPage {
					variables["memberCursor"] = query.Organization.Team.Members.PageInfo.EndCursor
				} else {
					break
				}
			}
		}

		repositories_detailed = make([]interface{}, 0, resultsPerPage) //removed this from the loop

		for {
			repository, resp, err := client.Teams.ListTeamReposByID(ctx, orgId, team.GetID(), &options.ListOptions)
			if err != nil {
				return err
			}

			for _, v := range repository {
				repositories = append(repositories, v.GetName())
				repositories_detailed = append(repositories_detailed, map[string]interface{}{
					"repo_id":   v.GetID(),
					"role_name": v.GetRoleName(),
				})
			}

			if resp.NextPage == 0 {
				break
			}
			options.Page = resp.NextPage
		}
	}

	d.SetId(strconv.FormatInt(team.GetID(), 10))
	if err = d.Set("name", team.GetName()); err != nil {
		return err
	}
	if err = d.Set("members", members); err != nil {
		return err
	}
	if err = d.Set("repositories", repositories); err != nil {
		return err
	}
	if err = d.Set("repositories_detailed", repositories_detailed); err != nil {
		return err
	}
	if err = d.Set("description", team.GetDescription()); err != nil {
		return err
	}
	if err = d.Set("privacy", team.GetPrivacy()); err != nil {
		return err
	}
	if err = d.Set("permission", team.GetPermission()); err != nil {
		return err
	}
	if err = d.Set("node_id", team.GetNodeID()); err != nil {
		return err
	}

	return nil
}
