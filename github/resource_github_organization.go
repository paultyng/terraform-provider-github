package github

import (
	"context"
	"log"
	"strconv"

	"github.com/google/go-github/v36/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubOrganizationCreate,
		Read:   resourceGithubOrganizationRead,
		Update: resourceGithubOrganizationUpdate,
		Delete: resourceGithubOrganizationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"login": {
				Type:     schema.TypeString,
				Required: true,
			},
			"admin": {
				Type:     schema.TypeString,
				Required: true,
			},
			"profile_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceGithubOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	ownerName := meta.(*Owner).name
	organizationName := github.String(d.Get("login").(string))
	enterpriseAdmin := github.String(d.Get("admin").(string))

	newOrganization := github.Organization{
		Login: organizationName,
	}

	ctx := context.Background()

	log.Printf("[DEBUG] Creating organization: %s (%s)", *organizationName, ownerName)
	githubOrganization, _, err := client.Admin.CreateOrg(ctx, &newOrganization, *enterpriseAdmin)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(githubOrganization.GetID(), 10))
	return resourceGithubOrganizationRead(d, meta)
}

func resourceGithubOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	organizationName := github.String(d.Get("login").(string))

	ctx := context.Background()

	log.Printf("[DEBUG] Reading organization: %s", *organizationName)
	org, _, err := client.Organizations.Get(ctx, *organizationName)
	if err != nil {
		return err
	}

	d.Set("login", org.GetLogin())
	d.Set("profile_name", org.GetName())

	return nil
}

func resourceGithubOrganizationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client

	ctx := context.Background()

	orgId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}

	oldOrg, _, err := client.Organizations.GetByID(ctx, orgId)
	if err != nil {
		return err
	}

	oldOrgName := *oldOrg.Login
	newOrgName := github.String(d.Get("login").(string))

	newOrganizationEdit := github.Organization{
		Login: &oldOrgName,
		Name:  github.String(d.Get("profile_name").(string)),
	}
	log.Printf("[DEBUG] Updating organization: %s", oldOrgName)
	org, _, err := client.Organizations.Edit(ctx, oldOrgName, &newOrganizationEdit)
	if err != nil {
		return err
	}

	// if d.HasChange("login") {
	if oldOrgName != *newOrgName {
		newOrganizationRename := github.Organization{
			Login: &oldOrgName,
		}

		log.Printf("[DEBUG] Renaming organization: %s -> %s", oldOrgName, *newOrgName)
		_, _, err = client.Admin.RenameOrg(ctx, &newOrganizationRename, *newOrgName)
		if err != nil {
			return err
		}
	}

	d.SetId(strconv.FormatInt(org.GetID(), 10))
	return resourceGithubOrganizationRead(d, meta)
}

func resourceGithubOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	// no github api for that?
	return nil
}
