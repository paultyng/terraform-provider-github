package github

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubTeamSyncGroupMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubTeamSyncGroupMappingCreate,
		Read:   resourceGithubTeamSyncGroupMappingRead,
		Update: resourceGithubTeamSyncGroupMappingUpdate,
		Delete: resourceGithubTeamSyncGroupMappingDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				if err := d.Set("team_slug", d.Id()); err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("teams/%s/team-sync/group-mappings", d.Id()))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"team_slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the team.",
			},
			"group": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "An Array of GitHub Identity Provider Groups (or empty []).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the IdP group.",
						},
						"group_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the IdP group.",
						},
						"group_description": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The description of the IdP group.",
						},
					},
				},
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubTeamSyncGroupMappingCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	ctx := context.Background()
	orgName := meta.(*Owner).name
	slug := d.Get("team_slug").(string)

	idpGroupList := expandTeamSyncGroups(d)
	_, _, err = client.Teams.CreateOrUpdateIDPGroupConnectionsBySlug(ctx, orgName, slug, *idpGroupList)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("teams/%s/team-sync/group-mappings", slug))

	return resourceGithubTeamSyncGroupMappingRead(d, meta)
}

func resourceGithubTeamSyncGroupMappingRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	slug := d.Get("team_slug").(string)

	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	idpGroupList, resp, err := client.Teams.ListIDPGroupsForTeamBySlug(ctx, orgName, slug)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing team_sync_group mapping for %s/%s from state because it no longer exists in GitHub",
					orgName, slug)
				d.SetId("")
				return nil
			}
		}
		return err
	}

	groups, err := flattenGithubIDPGroupList(idpGroupList)
	if err != nil {
		return err
	}

	if err = d.Set("group", groups); err != nil {
		return fmt.Errorf("error setting groups: %s", err)
	}
	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return fmt.Errorf("error setting etag: %s", err)
	}

	return nil
}

func resourceGithubTeamSyncGroupMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	slug := d.Get("team_slug").(string)

	idpGroupList := expandTeamSyncGroups(d)
	_, _, err = client.Teams.CreateOrUpdateIDPGroupConnectionsBySlug(ctx, orgName, slug, *idpGroupList)
	if err != nil {
		return err
	}

	return resourceGithubTeamSyncGroupMappingRead(d, meta)
}

func resourceGithubTeamSyncGroupMappingDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	slug := d.Get("team_slug").(string)

	groups := make([]*github.IDPGroup, 0)
	emptyGroupList := github.IDPGroupList{Groups: groups}

	_, _, err = client.Teams.CreateOrUpdateIDPGroupConnectionsBySlug(ctx, orgName, slug, emptyGroupList)
	return err
}

func flattenGithubIDPGroupList(idpGroupList *github.IDPGroupList) ([]interface{}, error) {
	if idpGroupList == nil {
		return make([]interface{}, 0), nil
	}
	results := make([]interface{}, 0)
	for _, group := range idpGroupList.Groups {
		result := make(map[string]interface{})
		result["group_id"] = group.GetGroupID()
		result["group_name"] = group.GetGroupName()
		result["group_description"] = group.GetGroupDescription()
		results = append(results, result)
	}

	return results, nil
}

// expandTeamSyncGroups creates an IDPGroupList with an array of IdP groups
// defined in the *schema.ResourceData to be later used to create or update
// IdP group connections in GitHub; if the "group" key is not present,
// an empty array must be set in the IDPGroupList per API endpoint specs:
// https://developer.github.com/v3/teams/team_sync/#create-or-update-idp-group-connections
func expandTeamSyncGroups(d *schema.ResourceData) *github.IDPGroupList {
	groups := make([]*github.IDPGroup, 0)
	if v, ok := d.GetOk("group"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			m := v.(map[string]interface{})
			groupID := m["group_id"].(string)
			groupName := m["group_name"].(string)
			groupDescription := m["group_description"].(string)
			group := &github.IDPGroup{
				GroupID:          github.String(groupID),
				GroupName:        github.String(groupName),
				GroupDescription: github.String(groupDescription),
			}
			groups = append(groups, group)
		}
	}
	return &github.IDPGroupList{Groups: groups}

}
