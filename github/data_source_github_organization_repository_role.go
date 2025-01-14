package github

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubOrganizationRepositoryRole() *schema.Resource {
	return &schema.Resource{
		Description: "Lookup a custom organization repository role.",

		Read: dataSourceGithubOrganizationRepositoryRoleRead,

		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "The ID of the organization repository role.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"name": {
				Description: "The name of the organization repository role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The description of the organization repository role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"base_role": {
				Description: "The system role from which this role inherits permissions.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"permissions": {
				Description: "The permissions included in this role.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
		},
	}
}

func dataSourceGithubOrganizationRepositoryRoleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleId := int64(d.Get("role_id").(int))

	// TODO: Use this code when go-github adds the functionality to get a custom repo role
	// role, _, err := client.Organizations.GetCustomRepoRole(ctx, orgName, roleId)
	// if err != nil {
	// 	return err
	// }

	roles, _, err := client.Organizations.ListCustomRepoRoles(ctx, orgName)
	if err != nil {
		return err
	}

	var role *github.CustomRepoRoles
	for _, r := range roles.CustomRepoRoles {
		if r.GetID() == roleId {
			role = r
			break
		}
	}
	if role == nil {
		return fmt.Errorf("custom organization repo role with ID %d not found", roleId)
	}

	r := map[string]any{
		"role_id":     role.GetID(),
		"name":        role.GetName(),
		"description": role.GetDescription(),
		"base_role":   role.GetBaseRole(),
		"permissions": role.Permissions,
	}

	d.SetId(strconv.FormatInt(role.GetID(), 10))

	for k, v := range r {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}
