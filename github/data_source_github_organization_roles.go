package github

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubOrganizationRoles() *schema.Resource {
	return &schema.Resource{
		Description: "Lookup all custom roles in an organization.",

		Read: dataSourceGithubOrganizationRolesRead,

		Schema: map[string]*schema.Schema{
			"roles": {
				Description: "Available organization roles.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_id": {
							Description: "The ID of the organization role.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"name": {
							Description: "The name of the organization role.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"description": {
							Description: "The description of the organization role.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"source": {
							Description: "The source of this role; one of `Predefined`, `Organization`, or `Enterprise`.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"base_role": {
							Description: "The system role from which this role inherits permissions.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"permissions": {
							Description: "A list of permissions included in this role.",
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubOrganizationRolesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	ret, _, err := client.Organizations.ListRoles(ctx, orgName)
	if err != nil {
		return err
	}

	allRoles := make([]any, ret.GetTotalCount())
	for i, role := range ret.CustomRepoRoles {
		r := map[string]any{
			"role_id":     role.GetID(),
			"name":        role.GetName(),
			"description": role.GetDescription(),
			"source":      role.GetSource(),
			"base_role":   role.GetBaseRole(),
			"permissions": role.Permissions,
		}
		allRoles[i] = r
	}

	d.SetId(fmt.Sprintf("%s/github-org-roles", orgName))
	if err := d.Set("roles", allRoles); err != nil {
		return fmt.Errorf("error setting roles: %s", err)
	}

	return nil
}
