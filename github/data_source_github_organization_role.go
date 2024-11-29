package github

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubOrganizationRole() *schema.Resource {
	return &schema.Resource{
		Description: "Lookup a custom organization role.",

		Read: dataSourceGithubOrganizationRoleRead,

		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "The ID of the organization role.",
				Type:        schema.TypeInt,
				Required:    true,
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
	}
}

func dataSourceGithubOrganizationRoleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleId := int64(d.Get("role_id").(int))

	role, _, err := client.Organizations.GetOrgRole(ctx, orgName, roleId)
	if err != nil {
		return err
	}

	r := map[string]any{
		"role_id":     role.GetID(),
		"name":        role.GetName(),
		"description": role.GetDescription(),
		"source":      role.GetSource(),
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
