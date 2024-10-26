package github

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/githubv4"
)

func resourceGithubTeam() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubTeamCreate,
		Read:   resourceGithubTeamRead,
		Update: resourceGithubTeamUpdate,
		Delete: resourceGithubTeamDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubTeamImport,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("slug", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChange("name")
			}),
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the team.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the team.",
			},
			"privacy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "secret",
				Description:      "The level of privacy for the team. Must be one of 'secret' or 'closed'.",
				ValidateDiagFunc: validateValueFunc([]string{"secret", "closed"}),
			},
			"parent_team_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The ID or slug of the parent team, if this is a nested team.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Get("parent_team_id") == d.Get("parent_team_read_id") || d.Get("parent_team_id") == d.Get("parent_team_read_slug") {
						return true
					}
					return false
				},
			},
			"parent_team_read_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The id of the parent team read in Github.",
			},
			"parent_team_read_slug": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The id of the parent team read in Github.",
			},
			"ldap_dn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The LDAP Distinguished Name of the group where membership will be synchronized. Only available in GitHub Enterprise Server.",
			},
			"create_default_maintainer": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Adds a default maintainer to the team. Adds the creating user to the team when 'true'.",
			},
			"slug": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The slug of the created team.",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Node ID of the created team.",
			},
			"members_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceGithubTeamCreate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	ownerName := meta.(*Owner).name
	name := d.Get("name").(string)

	newTeam := github.NewTeam{
		Name:        name,
		Description: github.String(d.Get("description").(string)),
		Privacy:     github.String(d.Get("privacy").(string)),
	}

	if ldapDN := d.Get("ldap_dn").(string); ldapDN != "" {
		newTeam.LDAPDN = &ldapDN
	}

	if parentTeamID, ok := d.GetOk("parent_team_id"); ok {
		teamId, err := getTeamID(parentTeamID.(string), meta)
		if err != nil {
			return err
		}
		newTeam.ParentTeamID = &teamId
	}
	ctx := context.Background()

	githubTeam, _, err := client.Teams.CreateTeam(ctx,
		ownerName, newTeam)
	if err != nil {
		return err
	}

	/*
		When using a GitHub App for authentication, `members:write` permissions on the App are needed.

		However, when using a GitHub App, CreateTeam will not correctly nest the team under the parent,
		if the parent team was created by someone else than the GitHub App. In that case, the response
		object will contain a `nil` parent object.

		This can be resolved by using an additional call to EditTeamByID. This will be able to set the
		parent team correctly when using a GitHub App with `members:write` permissions.

		Note that this is best-effort: when running this with a PAT that does not have admin permissions
		on the parent team, the operation might still fail to set the parent team.
	*/
	if newTeam.ParentTeamID != nil && githubTeam.Parent == nil {
		_, _, err := client.Teams.EditTeamByID(ctx,
			*githubTeam.Organization.ID,
			*githubTeam.ID,
			newTeam,
			false)

		if err != nil {
			return err
		}
	}

	create_default_maintainer := d.Get("create_default_maintainer").(bool)
	if !create_default_maintainer {
		log.Printf("[DEBUG] Removing default maintainer from team: %s (%s)", name, ownerName)
		if err := removeDefaultMaintainer(*githubTeam.Slug, meta); err != nil {
			return err
		}
	}

	d.SetId(strconv.FormatInt(githubTeam.GetID(), 10))
	return resourceGithubTeamRead(d, meta)
}

func resourceGithubTeamRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if !d.IsNewResource() {
		ctx = context.WithValue(ctx, ctxEtag, d.Get("etag").(string))
	}

	team, resp, err := client.Teams.GetTeamByID(ctx, orgId, id)
	if err != nil {
		if ghErr, ok := err.(*github.ErrorResponse); ok {
			if ghErr.Response.StatusCode == http.StatusNotModified {
				return nil
			}
			if ghErr.Response.StatusCode == http.StatusNotFound {
				log.Printf("[INFO] Removing team %s from state because it no longer exists in GitHub",
					d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("etag", resp.Header.Get("ETag")); err != nil {
		return err
	}
	if err = d.Set("description", team.GetDescription()); err != nil {
		return err
	}
	if err = d.Set("name", team.GetName()); err != nil {
		return err
	}
	if err = d.Set("privacy", team.GetPrivacy()); err != nil {
		return err
	}
	if parent := team.Parent; parent != nil {
		if err = d.Set("parent_team_id", strconv.FormatInt(team.Parent.GetID(), 10)); err != nil {
			return err
		}
		if err = d.Set("parent_team_read_id", strconv.FormatInt(team.Parent.GetID(), 10)); err != nil {
			return err
		}
		if err = d.Set("parent_team_read_slug", parent.Slug); err != nil {
			return err
		}
	} else {
		if err = d.Set("parent_team_id", ""); err != nil {
			return err
		}
		if err = d.Set("parent_team_read_id", ""); err != nil {
			return err
		}
		if err = d.Set("parent_team_read_slug", ""); err != nil {
			return err
		}
	}
	if err = d.Set("ldap_dn", team.GetLDAPDN()); err != nil {
		return err
	}
	if err = d.Set("slug", team.GetSlug()); err != nil {
		return err
	}
	if err = d.Set("node_id", team.GetNodeID()); err != nil {
		return err
	}
	if err = d.Set("members_count", team.GetMembersCount()); err != nil {
		return err
	}

	return nil
}

func resourceGithubTeamUpdate(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id
	var removeParentTeam bool

	editedTeam := github.NewTeam{
		Name:        d.Get("name").(string),
		Description: github.String(d.Get("description").(string)),
		Privacy:     github.String(d.Get("privacy").(string)),
	}
	if parentTeamID, ok := d.GetOk("parent_team_id"); ok {
		teamId, err := getTeamID(parentTeamID.(string), meta)
		if err != nil {
			return err
		}
		editedTeam.ParentTeamID = &teamId
		removeParentTeam = false
	} else {
		removeParentTeam = true
	}

	teamId, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	team, _, err := client.Teams.EditTeamByID(ctx, orgId, teamId, editedTeam, removeParentTeam)
	if err != nil {
		return err
	}

	if d.HasChange("ldap_dn") {
		ldapDN := d.Get("ldap_dn").(string)
		mapping := &github.TeamLDAPMapping{
			LDAPDN: github.String(ldapDN),
		}
		_, _, err = client.Admin.UpdateTeamLDAPMapping(ctx, team.GetID(), mapping)
		if err != nil {
			return err
		}
	}

	d.SetId(strconv.FormatInt(team.GetID(), 10))
	return resourceGithubTeamRead(d, meta)
}

func resourceGithubTeamDelete(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client
	orgId := meta.(*Owner).id

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return unconvertibleIdErr(d.Id(), err)
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	_, err = client.Teams.DeleteTeamByID(ctx, orgId, id)
	/*
		When deleting a team and it failed, we need to check if it has already been deleted meanwhile.
		This could be the case when deleting nested teams via Terraform by looping through a module
		or resource and the parent team might have been deleted already. If the parent team had
		been deleted already (via parallel runs), the child team is also already gone (deleted by
		GitHub automatically).
		So we're checking if it still exists and if not, simply remove it from TF state.
	*/
	if err != nil {
		// Fetch the team in order to see if it exists or not (http 404)
		_, _, err = client.Teams.GetTeamByID(ctx, orgId, id)
		if err != nil {
			if ghErr, ok := err.(*github.ErrorResponse); ok {
				if ghErr.Response.StatusCode == http.StatusNotFound {
					// If team we failed to delete does not exist, remove it from TF state.
					log.Printf("[WARN] Removing team: %s from state because it no longer exists",
						d.Id())
					d.SetId("")
					return nil
				}
			}
		}
	}
	return err
}

func resourceGithubTeamImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	teamId, err := getTeamID(d.Id(), meta)
	if err != nil {
		return nil, err
	}

	d.SetId(strconv.FormatInt(teamId, 10))
	if err = d.Set("create_default_maintainer", false); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func removeDefaultMaintainer(teamSlug string, meta interface{}) error {

	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	v4client := meta.(*Owner).v4client

	type User struct {
		Login githubv4.String
	}

	var query struct {
		Organization struct {
			Team struct {
				Members struct {
					Nodes []User
				}
			} `graphql:"team(slug:$slug)"`
		} `graphql:"organization(login:$login)"`
	}
	variables := map[string]interface{}{
		"slug":  githubv4.String(teamSlug),
		"login": githubv4.String(orgName),
	}

	err := v4client.Query(meta.(*Owner).StopContext, &query, variables)
	if err != nil {
		return err
	}

	for _, user := range query.Organization.Team.Members.Nodes {
		_, err := client.Teams.RemoveTeamMembershipBySlug(meta.(*Owner).StopContext, orgName, teamSlug, string(user.Login))
		if err != nil {
			return err
		}
	}

	return nil
}
