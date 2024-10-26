package github

import (
	"context"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/githubv4"
)

type MemberChange struct {
	Old, New map[string]interface{}
}

func resourceGithubTeamMembers() *schema.Resource {

	return &schema.Resource{
		Create: resourceGithubTeamMembersCreate,
		Read:   resourceGithubTeamMembersRead,
		Update: resourceGithubTeamMembersUpdate,
		Delete: resourceGithubTeamMembersDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubTeamMembersImport,
		},

		Schema: map[string]*schema.Schema{
			"team_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub team id or slug",
			},
			"members": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of team members.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: caseInsensitive(),
							Description:      "The user to add to the team.",
						},
						"role": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "member",
							Description:      "The role of the user within the team. Must be one of 'member' or 'maintainer'.",
							ValidateDiagFunc: validateValueFunc([]string{"member", "maintainer"}),
						},
					},
				},
			},
		},
	}
}

func resourceGithubTeamMembersCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id

	teamIdString := d.Get("team_id").(string)
	teamId, err := getTeamID(teamIdString, meta)
	if err != nil {
		return err
	}
	ctx := context.Background()

	members := d.Get("members").(*schema.Set)
	for _, mMap := range members.List() {
		memb := mMap.(map[string]interface{})
		username := memb["username"].(string)
		role := memb["role"].(string)

		log.Printf("[DEBUG] Creating team membership: %s/%s (%s)", teamIdString, username, role)
		_, _, err = client.Teams.AddTeamMembershipByID(ctx,
			orgId,
			teamId,
			username,
			&github.TeamAddTeamMembershipOptions{
				Role: role,
			},
		)
		if err != nil {
			return err
		}
	}

	d.SetId(teamIdString)

	return resourceGithubTeamMembersRead(d, meta)
}

func resourceGithubTeamMembersUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id

	teamIdString := d.Get("team_id").(string)
	teamId, err := getTeamID(teamIdString, meta)
	if err != nil {
		return err
	}
	ctx := context.Background()

	o, n := d.GetChange("members")
	vals := make(map[string]*MemberChange)
	for _, raw := range o.(*schema.Set).List() {
		obj := raw.(map[string]interface{})
		k := obj["username"].(string)
		vals[k] = &MemberChange{Old: obj}
	}
	for _, raw := range n.(*schema.Set).List() {
		obj := raw.(map[string]interface{})
		k := obj["username"].(string)
		if _, ok := vals[k]; !ok {
			vals[k] = &MemberChange{}
		}
		vals[k].New = obj
	}

	for username, change := range vals {
		var create, delete bool

		switch {
		// create a new one if old is nil
		case change.Old == nil:
			create = true
		// delete existing if new is nil
		case change.New == nil:
			delete = true
			// no change
		case reflect.DeepEqual(change.Old, change.New):
			continue
			// recreate - role changed
		default:
			delete = true
			create = true
		}

		if delete {
			log.Printf("[DEBUG] Deleting team membership: %s/%s", teamIdString, username)

			_, err = client.Teams.RemoveTeamMembershipByID(ctx, orgId, teamId, username)
			if err != nil {
				return err
			}
		}

		if create {
			role := change.New["role"].(string)

			log.Printf("[DEBUG] Creating team membership: %s/%s (%s)", teamIdString, username, role)
			_, _, err = client.Teams.AddTeamMembershipByID(ctx,
				orgId,
				teamId,
				username,
				&github.TeamAddTeamMembershipOptions{
					Role: role,
				},
			)
			if err != nil {
				return err
			}
		}
	}

	d.SetId(teamIdString)

	return resourceGithubTeamMembersRead(d, meta)
}

func resourceGithubTeamMembersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v4client
	orgName := meta.(*Owner).name
	teamIdString := d.Get("team_id").(string)
	if teamIdString == "" && !d.IsNewResource() {
		log.Printf("[DEBUG] Importing team with id %q", d.Id())
		teamIdString = d.Id()
	}

	teamSlug, err := getTeamSlug(teamIdString, meta)
	if err != nil {
		return err
	}

	// We intentionally set these early to allow reconciliation
	// from an upstream bug which emptied team_id in state
	// See https://github.com/integrations/terraform-provider-github/issues/323
	if err := d.Set("team_id", teamIdString); err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	log.Printf("[DEBUG] Reading team members: %s", teamIdString)
	var q struct {
		Organization struct {
			Team struct {
				Members struct {
					Edges []struct {
						Node struct {
							Login string
						}
						Role string
					}
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"members(membership:IMMEDIATE, first:100, after: $after)"`
			} `graphql:"team(slug:$teamSlug)"`
		} `graphql:"organization(login:$orgName)"`
	}

	variables := map[string]interface{}{
		"teamSlug": githubv4.String(teamSlug),
		"orgName":  githubv4.String(orgName),
		"after":    (*githubv4.String)(nil),
	}

	var teamMembersAndMaintainers []interface{}
	for {
		if err := client.Query(ctx, &q, variables); err != nil {
			return err
		}

		// Add all members to the list
		for _, member := range q.Organization.Team.Members.Edges {
			teamMembersAndMaintainers = append(teamMembersAndMaintainers, map[string]interface{}{
				"username": member.Node.Login,
				"role":     strings.ToLower(member.Role),
			})
		}
		if !q.Organization.Team.Members.PageInfo.HasNextPage {
			break
		}
		variables["after"] = githubv4.NewString(q.Organization.Team.Members.PageInfo.EndCursor)
	}

	if err := d.Set("members", teamMembersAndMaintainers); err != nil {
		return err
	}

	return nil
}

func resourceGithubTeamMembersDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id
	teamIdString := d.Get("team_id").(string)
	teamId, err := getTeamID(teamIdString, meta)
	if err != nil {
		return err
	}

	members := d.Get("members").(*schema.Set)
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	for _, member := range members.List() {
		mem := member.(map[string]interface{})
		username := mem["username"].(string)

		log.Printf("[DEBUG] Deleting team membership: %s/%s", teamIdString, username)

		_, err = client.Teams.RemoveTeamMembershipByID(ctx, orgId, teamId, username)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceGithubTeamMembersImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	teamId, err := getTeamID(d.Id(), meta)
	if err != nil {
		return nil, err
	}

	d.SetId(strconv.FormatInt(teamId, 10))

	return []*schema.ResourceData{d}, nil
}
