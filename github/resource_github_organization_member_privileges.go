package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v32/github"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGithubOrganizationMemberPrivileges() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubOrganizationMemberPrivilegesCreate,
		Read:   resourceGithubOrganizationMemberPrivilegesRead,
		Update: resourceGithubOrganizationMemberPrivilegesUpdate,
		Delete: resourceGithubOrganizationMemberPrivilegesDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"default_repository_permission": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "read",
				ValidateFunc: validateValueFunc([]string{"read", "write", "admin", "none"}),
			},
			"members_can_create_repositories": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"members_can_create_internal_repositories": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"members_can_create_private_repositories": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"members_can_create_public_repositories": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func updateGithubOrganizationMemberPrivileges(d *schema.ResourceData, meta interface{}) (int64, error) {
	err := checkOrganization(meta)
	if err != nil {
		return 0, err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	defaultRepoPermission := d.Get("default_repository_permission").(string)
	membersCanCreateRepos := d.Get("members_can_create_repositories").(bool)
	membersCanCreatePublicRepos := d.Get("members_can_create_public_repositories").(bool)
	membersCanCreatePrivateRepos := d.Get("members_can_create_private_repositories").(bool)
	ctx := context.Background()

	organization, _, err := client.Organizations.Get(ctx, orgName)
	if err != nil {
		return 0, err
	}

	plan := organization.GetPlan()

	log.Printf("[DEBUG] Updating organization member privileges %s", orgName)
	if *plan.Name == "free" {
		organization, _, err = client.Organizations.Edit(ctx, orgName, &github.Organization{
			DefaultRepoPermission:        &defaultRepoPermission,
			MembersCanCreateRepos:        &membersCanCreateRepos,
			MembersCanCreatePublicRepos:  &membersCanCreatePublicRepos,
			MembersCanCreatePrivateRepos: &membersCanCreatePrivateRepos,
		})
	} else {
		membersCanCreateInternalRepos := d.Get("members_can_create_internal_repositories").(bool)
		organization, _, err = client.Organizations.Edit(ctx, orgName, &github.Organization{
			DefaultRepoPermission:         &defaultRepoPermission,
			MembersCanCreateRepos:         &membersCanCreateRepos,
			MembersCanCreatePublicRepos:   &membersCanCreatePublicRepos,
			MembersCanCreatePrivateRepos:  &membersCanCreatePrivateRepos,
			MembersCanCreateInternalRepos: &membersCanCreateInternalRepos,
		})
	}
	if err != nil {
		return 0, err
	}
	return organization.GetID(), err
}

func resourceGithubOrganizationMemberPrivilegesCreate(d *schema.ResourceData, meta interface{}) error {
	id, err := updateGithubOrganizationMemberPrivileges(d, meta)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(id, 10))

	return resourceGithubOrganizationMemberPrivilegesRead(d, meta)
}

func resourceGithubOrganizationMemberPrivilegesRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	organizationID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	log.Printf("[DEBUG] Reading organization: %s", orgName)
	organization, resp, err := client.Organizations.GetByID(ctx, organizationID)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[WARN] Removing organization member privileges %s from state because organization %s no longer exists in GitHub",
					d.Id(), orgName)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("etag", resp.Header.Get("ETag"))
	d.Set("default_repository_permission", organization.GetDefaultRepoPermission())
	d.Set("members_can_create_repositories", organization.GetMembersCanCreateRepos())
	d.Set("members_can_create_public_repositories", organization.GetMembersCanCreatePublicRepos())
	d.Set("members_can_create_private_repositories", organization.GetMembersCanCreatePrivateRepos())
	d.Set("members_can_create_internal_repositories", organization.GetMembersCanCreateInternalRepos())

	return nil
}

func resourceGithubOrganizationMemberPrivilegesUpdate(d *schema.ResourceData, meta interface{}) error {
	_, err := updateGithubOrganizationMemberPrivileges(d, meta)
	if err != nil {
		return err
	}

	return resourceGithubOrganizationMemberPrivilegesRead(d, meta)
}

func resourceGithubOrganizationMemberPrivilegesDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")

	return nil
}
