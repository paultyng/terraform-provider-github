package github

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubEMUGroupMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubEMUGroupMappingCreate,
		Read:   resourceGithubEMUGroupMappingRead,
		Update: resourceGithubEMUGroupMappingUpdate,
		Delete: resourceGithubEMUGroupMappingDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				id, err := strconv.Atoi(d.Id())
				if err != nil {
					return nil, err
				}
				if err := d.Set("group_id", id); err != nil {
					return nil, err
				}
				ctx := context.WithValue(context.Background(), ctxId, d.Id())
				client := meta.(*Owner).v3client
				orgName := meta.(*Owner).name
				group, _, err := client.Teams.GetExternalGroup(ctx, orgName, int64(id))
				if err != nil {
					return nil, err
				}
				if len(group.Teams) != 1 {
					return nil, fmt.Errorf("could not get team_slug from %v number of teams", len(group.Teams))
				}
				if err := d.Set("team_slug", group.Teams[0].TeamName); err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("teams/%s/external-groups", d.Id()))
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"team_slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Slug of the GitHub team.",
			},
			"group_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Integer corresponding to the external group ID to be linked.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGithubEMUGroupMappingCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceGithubEMUGroupMappingUpdate(d, meta)
}

func resourceGithubEMUGroupMappingRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	id, ok := d.GetOk("group_id")
	if !ok {
		return fmt.Errorf("could not get group id from provided value")
	}
	id64, err := getInt64FromInterface(id)
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	group, resp, err := client.Teams.GetExternalGroup(ctx, orgName, id64)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			// If the group is not found, remove it from state
			d.SetId("")
			return nil
		}
		return err
	}

	if len(group.Teams) < 1 {
		// if there's not a team linked, that means it was removed outside of terraform
		// and we should remove it from our state
		d.SetId("")
		return nil
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("group_id", int(*group.GroupID)); err != nil {
		return err
	}
	return nil
}

func resourceGithubEMUGroupMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	teamSlug, ok := d.GetOk("team_slug")
	if !ok {
		return fmt.Errorf("could not get team slug from provided value")
	}

	id, ok := d.GetOk("group_id")
	if !ok {
		return fmt.Errorf("could not get group id from provided value")
	}
	id64, err := getInt64FromInterface(id)
	if err != nil {
		return err
	}

	eg := &github.ExternalGroup{
		GroupID: &id64,
	}

	_, _, err = client.Teams.UpdateConnectedExternalGroup(ctx, orgName, teamSlug.(string), eg)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("teams/%s/external-groups", teamSlug))
	return resourceGithubEMUGroupMappingRead(d, meta)
}

func resourceGithubEMUGroupMappingDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name

	teamSlug, ok := d.GetOk("team_slug")
	if !ok {
		return fmt.Errorf("could not parse team slug from provided value")
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Teams.RemoveConnectedExternalGroup(ctx, orgName, teamSlug.(string))
	if err != nil {
		return err
	}
	return nil
}

func getInt64FromInterface(val interface{}) (int64, error) {
	var id64 int64
	switch val := val.(type) {
	case int64:
		id64 = val
	case int:
		id64 = int64(val)
	case string:
		var err error
		id64, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("could not parse id from string: %v", err)
		}
	default:
		return 0, fmt.Errorf("unexpected type converting to int64 from interface")
	}
	return id64, nil
}
