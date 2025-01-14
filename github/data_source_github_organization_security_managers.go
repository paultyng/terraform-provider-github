package github

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubOrganizationSecurityManagers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubOrganizationSecurityManagersRead,

		Schema: map[string]*schema.Schema{
			"teams": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Unique identifier of the team.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"slug": {
							Description: "Name based identifier of the team.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Name of the team.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"permission": {
							Description: "Permission that the team will have for its repositories.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubOrganizationSecurityManagersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	allTeams := make([]interface{}, 0)

	teams, _, err := client.Organizations.ListSecurityManagerTeams(ctx, orgName)
	if err != nil {
		return err
	}

	for _, team := range teams {
		t := map[string]any{
			"id":         team.GetID(),
			"slug":       team.GetSlug(),
			"name":       team.GetName(),
			"permission": team.GetPermission(),
		}
		allTeams = append(allTeams, t)
	}

	d.SetId(fmt.Sprintf("%s/github-org-security-managers", orgName))
	if err := d.Set("teams", allTeams); err != nil {
		return fmt.Errorf("error setting teams: %s", err)
	}

	return nil
}
