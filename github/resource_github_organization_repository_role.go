package github

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubOrganizationRepositoryRole() *schema.Resource {
	return &schema.Resource{
		Description: "Manage a custom organization repository role.",
		Create:      resourceGithubOrganizationRepositoryRoleCreate,
		Read:        resourceGithubOrganizationRepositoryRoleRead,
		Update:      resourceGithubOrganizationRepositoryRoleUpdate,
		Delete:      resourceGithubOrganizationRepositoryRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "The ID of the organization repository role.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"name": {
				Description: "The name of the organization repository role.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of the organization repository role.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"base_role": {
				Description:      "The base role for the organization repository role.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateValueFunc([]string{"read", "triage", "write", "maintain"}),
			},
			"permissions": {
				Description: "The permissions for the organization repository role.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
			},
		},
	}
}

func resourceGithubOrganizationRepositoryRoleCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.Background()

	permissions := d.Get("permissions").(*schema.Set).List()
	permissionsStr := make([]string, len(permissions))
	for i, v := range permissions {
		permissionsStr[i] = v.(string)
	}

	role, _, err := client.Organizations.CreateCustomRepoRole(ctx, orgName, &github.CreateOrUpdateCustomRepoRoleOptions{
		Name:        github.String(d.Get("name").(string)),
		Description: github.String(d.Get("description").(string)),
		BaseRole:    github.String(d.Get("base_role").(string)),
		Permissions: permissionsStr,
	})
	if err != nil {
		return fmt.Errorf("error creating GitHub organization repository role (%s/%s): %s", orgName, d.Get("name").(string), err)
	}

	d.SetId(fmt.Sprint(role.GetID()))
	return resourceGithubOrganizationRoleRead(d, meta)
}

func resourceGithubOrganizationRepositoryRoleRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	// TODO: Use this code when go-github adds the functionality to get a custom repo role
	// role, _, err := client.Organizations.GetCustomRepoRole(ctx, orgName, roleId)
	// if err != nil {
	// 	if ghErr, ok := err.(*github.ErrorResponse); ok {
	// 		if ghErr.Response.StatusCode == http.StatusNotFound {
	// 			log.Printf("[WARN] GitHub organization repository role (%s/%d) not found, removing from state", orgName, roleId)
	// 			d.SetId("")
	// 			return nil
	// 		}
	// 	}
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
		log.Printf("[WARN] GitHub organization repository role (%s/%d) not found, removing from state", orgName, roleId)
		d.SetId("")
		return nil
	}

	if err = d.Set("role_id", role.GetID()); err != nil {
		return err
	}
	if err = d.Set("name", role.Name); err != nil {
		return err
	}
	if err = d.Set("description", role.Description); err != nil {
		return err
	}
	if err = d.Set("base_role", role.BaseRole); err != nil {
		return err
	}
	if err = d.Set("permissions", role.Permissions); err != nil {
		return err
	}

	return nil
}

func resourceGithubOrganizationRepositoryRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	permissions := d.Get("permissions").(*schema.Set).List()
	permissionsStr := make([]string, len(permissions))
	for i, v := range permissions {
		permissionsStr[i] = v.(string)
	}

	update := &github.CreateOrUpdateCustomRepoRoleOptions{
		Name:        github.String(d.Get("name").(string)),
		Description: github.String(d.Get("description").(string)),
		BaseRole:    github.String(d.Get("base_role").(string)),
		Permissions: permissionsStr,
	}

	_, _, err = client.Organizations.UpdateCustomRepoRole(ctx, orgName, roleId, update)
	if err != nil {
		return fmt.Errorf("error updating GitHub organization repository role (%s/%s): %s", orgName, d.Get("name").(string), err)
	}

	return resourceGithubOrganizationRoleRead(d, meta)
}

func resourceGithubOrganizationRepositoryRoleDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name

	roleId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	_, err = client.Organizations.DeleteCustomRepoRole(ctx, orgName, roleId)
	if err != nil {
		return fmt.Errorf("Error deleting GitHub organization repository role %s (%d): %s", orgName, roleId, err)
	}

	return nil
}
