package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubOrganizationRole() *schema.Resource {
	return &schema.Resource{
		Description: "Manage a custom organization role.",

		Create: resourceGithubOrganizationRoleCreate,
		Read:   resourceGithubOrganizationRoleRead,
		Update: resourceGithubOrganizationRoleUpdate,
		Delete: resourceGithubOrganizationRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "The ID of the organization role.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"name": {
				Description: "The name of the organization role.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of the organization role.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"base_role": {
				Description:      "The base role for the organization role.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validateValueFunc([]string{"read", "triage", "write", "maintain", "admin"}),
			},
			"permissions": {
				Description: "The permissions for the organization role.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
			},
		},
	}
}

func resourceGithubOrganizationRoleCreate(d *schema.ResourceData, meta interface{}) error {
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

	role, _, err := client.Organizations.CreateCustomOrgRole(ctx, orgName, &github.CreateOrUpdateOrgRoleOptions{
		Name:        github.String(d.Get("name").(string)),
		Description: github.String(d.Get("description").(string)),
		BaseRole:    github.String(d.Get("base_role").(string)),
		Permissions: permissionsStr,
	})
	if err != nil {
		return fmt.Errorf("error creating GitHub custom organization role (%s/%s): %s", orgName, d.Get("name").(string), err)
	}

	d.SetId(fmt.Sprint(role.GetID()))
	return resourceGithubOrganizationRoleRead(d, meta)
}

func resourceGithubOrganizationRoleRead(d *schema.ResourceData, meta interface{}) error {
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

	role, _, err := client.Organizations.GetOrgRole(ctx, orgName, roleId)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] GitHub custom organization role (%s/%d) not found, removing from state", orgName, roleId)
				d.SetId("")
				return nil
			}
		}
		return err
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

func resourceGithubOrganizationRoleUpdate(d *schema.ResourceData, meta interface{}) error {
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

	update := &github.CreateOrUpdateOrgRoleOptions{
		Name:        github.String(d.Get("name").(string)),
		Description: github.String(d.Get("description").(string)),
		BaseRole:    github.String(d.Get("base_role").(string)),
		Permissions: permissionsStr,
	}

	_, _, err = client.Organizations.UpdateCustomOrgRole(ctx, orgName, roleId, update)
	if err != nil {
		return fmt.Errorf("error updating GitHub custom organization role (%s/%s): %s", orgName, d.Get("name").(string), err)
	}

	return resourceGithubOrganizationRoleRead(d, meta)
}

func resourceGithubOrganizationRoleDelete(d *schema.ResourceData, meta interface{}) error {
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

	_, err = client.Organizations.DeleteCustomOrgRole(ctx, orgName, roleId)
	if err != nil {
		return fmt.Errorf("Error deleting GitHub custom organization role %s (%d): %s", orgName, roleId, err)
	}

	return nil
}
