package github

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v39/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceGithubOrganizationTeamSyncGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubOrganizationTeamSyncGroupsRead,

		Schema: map[string]*schema.Schema{
			"groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubOrganizationTeamSyncGroupsRead(d *schema.ResourceData, meta interface{}) error {
	log.Print("[INFO] Refreshing GitHub Organization Team-Sync Groups")

	client := meta.(*Owner).v3client
	ctx := context.Background()

	orgName := meta.(*Owner).name
	options := &github.ListCursorOptions{PerPage: maxPerPage}

	groups := make([]interface{}, 0)
	for {
		idpGroupList, resp, err := client.Teams.ListIDPGroupsInOrganization(ctx, orgName, options)
		if err != nil {
			return err
		}

		result, err := flattenGithubIDPGroupList(idpGroupList)
		if err != nil {
			return fmt.Errorf("unable to flatten IdP Groups in Github Organization(Org: %q) : %+v", orgName, err)
		}

		groups = append(groups, result...)

		if resp.NextPageToken == "" {
			break
		}
		options.Page = resp.NextPageToken
	}

	d.SetId(fmt.Sprintf("%s/github-org-team-sync-groups", orgName))
	if err := d.Set("groups", groups); err != nil {
		return fmt.Errorf("error setting groups: %s", err)
	}

	return nil
}
