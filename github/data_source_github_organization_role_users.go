package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubOrganizationRoleUsers() *schema.Resource {
	return &schema.Resource{
		Description: "Lookup all users assigned to a custom organization role.",

		Read: dataSourceGithubOrganizationRoleUsersRead,

		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "The ID of the organization role.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"users": {
				Description: "Users assigned to the organization role.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_id": {
							Description: "The ID of the user.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"login": {
							Description: "The login for the GitHub user account.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						// See https://github.com/google/go-github/issues/3364
						// "assignment": {
						// 	Description: "Determines if the team has a direct, indirect, or mixed relationship to a role.",
						// 	Type:        schema.TypeString,
						// 	Computed:    true,
						// },
						// "parent_team_id": {
						// 	Description: "The ID of the parent team if this is an indirect assignment.",
						// 	Type:        schema.TypeString,
						// 	Computed:    true,
						// },
						// "parent_team_slug": {
						// 	Description: "The slug of the parent team if this is an indirect assignment.",
						// 	Type:        schema.TypeString,
						// 	Computed:    true,
						// },
					},
				},
			},
		},
	}
}

func dataSourceGithubOrganizationRoleUsersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleId := int64(d.Get("role_id").(int))

	allUsers := make([]any, 0)

	opts := &github.ListOptions{
		PerPage: maxPerPage,
	}

	for {
		users, resp, err := client.Organizations.ListUsersAssignedToOrgRole(ctx, orgName, roleId, opts)
		if err != nil {
			return err
		}

		for _, user := range users {
			u := map[string]any{
				"id":    user.GetID(),
				"login": user.GetLogin(),
			}
			allUsers = append(allUsers, u)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	d.SetId(fmt.Sprintf("%d", roleId))
	if err := d.Set("users", allUsers); err != nil {
		return fmt.Errorf("error setting users: %s", err)
	}

	return nil
}
