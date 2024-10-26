package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubExternalGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubExternalGroupsRead,
		Schema: map[string]*schema.Schema{
			"external_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubExternalGroupsRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	opts := &github.ListExternalGroupsOptions{}

	externalGroups := new(github.ExternalGroupList)

	for {
		groups, resp, err := client.Teams.ListExternalGroups(ctx, orgName, opts)
		if err != nil {
			return err
		}

		externalGroups.Groups = append(externalGroups.Groups, groups.Groups...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// convert to JSON in order to martial to format we can return
	jsonGroups, err := json.Marshal(externalGroups.Groups)
	if err != nil {
		return err
	}

	groupsState := make([]map[string]interface{}, 0)
	err = json.Unmarshal(jsonGroups, &groupsState)
	if err != nil {
		return err
	}

	if err := d.Set("external_groups", groupsState); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("/orgs/%v/external-groups", orgName))
	return nil
}
