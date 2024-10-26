package github

import (
	"context"

	"github.com/google/go-github/v65/github"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubActionsOrganizationVariables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubActionsOrganizationVariablesRead,

		Schema: map[string]*schema.Schema{
			"variables": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"visibility": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
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

func dataSourceGithubActionsOrganizationVariablesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name

	options := github.ListOptions{
		PerPage: 100,
	}

	var all_variables []map[string]string
	for {
		variables, resp, err := client.Actions.ListOrgVariables(context.TODO(), owner, &options)
		if err != nil {
			return err
		}
		for _, variable := range variables.Variables {
			new_variable := map[string]string{
				"name":       variable.Name,
				"value":      variable.Value,
				"visibility": *variable.Visibility,
				"created_at": variable.CreatedAt.String(),
				"updated_at": variable.UpdatedAt.String(),
			}
			all_variables = append(all_variables, new_variable)
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	d.SetId(owner)
	err := d.Set("variables", all_variables)
	if err != nil {
		return err
	}

	return nil
}
