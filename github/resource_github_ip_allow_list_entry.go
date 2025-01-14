package github

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/githubv4"
)

type IPAllowListEntry struct {
	AllowListValue githubv4.String
	CreatedAt      githubv4.DateTime
	ID             githubv4.ID
	IsActive       githubv4.Boolean
	Name           *githubv4.String
	UpdatedAt      githubv4.DateTime
}

func resourceGithubIPAllowListEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubIPAllowListEntryCreate,
		Read:   resourceGithubIPAllowListEntryRead,
		Update: resourceGithubIPAllowListEntryUpdate,
		Delete: resourceGithubIPAllowListEntryDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubIPAllowListEntryImport,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifies the date and time when the object was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifies the date and time when the object was last updated.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A single IP address or range of IP addresses in CIDR notation.",
			},
			"active": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Whether the entry is currently active.",
			},
			"owner": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The GraphQL ID of the owner (an Enterprise, Organization or App) for which to create the new IP allow list entry.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional name for the IP allow list entry.",
			},
		},
	}
}

func resourceGithubIPAllowListEntryCreate(d *schema.ResourceData, meta interface{}) error {
	var mutate struct {
		CreateIPAllowListEntry struct {
			IPAllowListEntry IPAllowListEntry `graphql:"ipAllowListEntry"`
		} `graphql:"createIpAllowListEntry(input:$input)"`
	}

	input := githubv4.CreateIpAllowListEntryInput{
		OwnerID:        githubv4.ID(d.Get("owner").(string)),
		AllowListValue: githubv4.String(d.Get("value").(string)),
		IsActive:       githubv4.Boolean(d.Get("active").(bool)),
		Name:           githubv4.NewString(githubv4.String(d.Get("name").(string))),
	}

	client := meta.(*Owner).v4client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	if err := client.Mutate(ctx, &mutate, input, nil); err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s", mutate.CreateIPAllowListEntry.IPAllowListEntry.ID))
	if err := d.Set("created_at", mutate.CreateIPAllowListEntry.IPAllowListEntry.CreatedAt.String()); err != nil {
		return err
	}
	if err := d.Set("updated_at", mutate.CreateIPAllowListEntry.IPAllowListEntry.UpdatedAt.String()); err != nil {
		return err
	}

	return nil
}

func resourceGithubIPAllowListEntryRead(d *schema.ResourceData, meta interface{}) error {
	var query struct {
		Node struct {
			// NOTE: We intentionally do not fetch/update the 'owner' attribute as it is immutable
			// This would require additional permissions on the GitHub token (ex: enterprise:admin)
			Node IPAllowListEntry `graphql:"... on IpAllowListEntry"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": d.Id(),
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	client := meta.(*Owner).v4client
	if err := client.Query(ctx, &query, variables); err != nil {
		if strings.Contains(err.Error(), "Could not resolve to a node with the global id") {
			log.Printf("[INFO] Removing IpAllowListEntry (%s) from state because it no longer exists in GitHub", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	if err := d.Set("value", string(query.Node.Node.AllowListValue)); err != nil {
		return err
	}
	if err := d.Set("active", bool(query.Node.Node.IsActive)); err != nil {
		return err
	}
	if query.Node.Node.Name != nil {
		if err := d.Set("name", string(*query.Node.Node.Name)); err != nil {
			return err
		}
	} else {
		if err := d.Set("name", ""); err != nil {
			return err
		}
	}
	if err := d.Set("created_at", query.Node.Node.CreatedAt.String()); err != nil {
		return err
	}
	if err := d.Set("updated_at", query.Node.Node.UpdatedAt.String()); err != nil {
		return err
	}

	return nil
}

func resourceGithubIPAllowListEntryUpdate(d *schema.ResourceData, meta interface{}) error {
	var mutate struct {
		CreateIPAllowListEntry struct {
			IPAllowListEntry IPAllowListEntry `graphql:"ipAllowListEntry"`
		} `graphql:"updateIpAllowListEntry(input:$input)"`
	}

	input := githubv4.UpdateIpAllowListEntryInput{
		IPAllowListEntryID: githubv4.ID(d.Id()),
		AllowListValue:     githubv4.String(d.Get("value").(string)),
		IsActive:           githubv4.Boolean(d.Get("active").(bool)),
		Name:               githubv4.NewString(githubv4.String(d.Get("name").(string))),
	}

	client := meta.(*Owner).v4client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	err := client.Mutate(ctx, &mutate, input, nil)
	if err != nil {
		return err
	}
	if err := d.Set("updated_at", mutate.CreateIPAllowListEntry.IPAllowListEntry.UpdatedAt.String()); err != nil {
		return err
	}
	return nil
}

func resourceGithubIPAllowListEntryDelete(d *schema.ResourceData, meta interface{}) error {
	var mutate struct {
		CreateIPAllowListEntry struct {
			IPAllowListEntry IPAllowListEntry `graphql:"ipAllowListEntry"`
		} `graphql:"deleteIpAllowListEntry(input:$input)"`
	}

	input := githubv4.DeleteIpAllowListEntryInput{
		IPAllowListEntryID: githubv4.ID(d.Id()),
	}

	client := meta.(*Owner).v4client
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	err := client.Mutate(ctx, &mutate, input, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceGithubIPAllowListEntryImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// We need to determine the type of the owner union so we can query for _just_ that type
	var query struct {
		Node struct {
			Node struct {
				Owner struct {
					Type githubv4.String `graphql:"__typename"`
				} `graphql:"owner"`
			} `graphql:"... on IpAllowListEntry"`
		} `graphql:"node(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": d.Id(),
	}
	ctx := context.WithValue(context.Background(), ctxId, d.Id())
	client := meta.(*Owner).v4client
	if err := client.Query(ctx, &query, variables); err != nil {
		return []*schema.ResourceData{d}, err
	}
	// Based on the type of owner resource, query the ID of that resource type
	// This allows us to only query for resource types that are actually in use
	// This is critical because we can hit permissions errors from GitHub otherwise
	// For example, '... on Enterprise' requires the 'enterprise:admin' scope
	switch query.Node.Node.Owner.Type {
	case "Organization":
		var query struct {
			Node struct {
				Node struct {
					Owner struct {
						Organization struct {
							ID githubv4.ID
						} `graphql:"... on Organization"`
					} `graphql:"owner"`
				} `graphql:"... on IpAllowListEntry"`
			} `graphql:"node(id: $id)"`
		}
		if err := client.Query(ctx, &query, variables); err != nil {
			return []*schema.ResourceData{d}, err
		}
		if err := d.Set("owner", fmt.Sprintf("%s", query.Node.Node.Owner.Organization.ID)); err != nil {
			return []*schema.ResourceData{d}, err
		}
	case "Enterprise":
		var query struct {
			Node struct {
				Node struct {
					Owner struct {
						Enterprise struct {
							ID githubv4.ID
						} `graphql:"... on Enterprise"`
					} `graphql:"owner"`
				} `graphql:"... on IpAllowListEntry"`
			} `graphql:"node(id: $id)"`
		}
		if err := client.Query(ctx, &query, variables); err != nil {
			return []*schema.ResourceData{d}, err
		}
		if err := d.Set("owner", fmt.Sprintf("%s", query.Node.Node.Owner.Enterprise.ID)); err != nil {
			return []*schema.ResourceData{d}, err
		}
	case "App":
		var query struct {
			Node struct {
				Node struct {
					Owner struct {
						App struct {
							ID githubv4.ID
						} `graphql:"... on App"`
					} `graphql:"owner"`
				} `graphql:"... on IpAllowListEntry"`
			} `graphql:"node(id: $id)"`
		}
		if err := client.Query(ctx, &query, variables); err != nil {
			return []*schema.ResourceData{d}, err
		}
		if err := d.Set("owner", fmt.Sprintf("%s", query.Node.Node.Owner.App.ID)); err != nil {
			return []*schema.ResourceData{d}, err
		}
	default:
		return []*schema.ResourceData{d}, fmt.Errorf("unexpected owner type: %q", query.Node.Node.Owner.Type)
	}
	return []*schema.ResourceData{d}, nil
}
