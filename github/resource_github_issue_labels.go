package github

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubIssueLabels() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubIssueLabelsCreateOrUpdate,
		Read:   resourceGithubIssueLabelsRead,
		Update: resourceGithubIssueLabelsCreateOrUpdate,
		Delete: resourceGithubIssueLabelsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub repository.",
			},
			"label": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of labels",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the label.",
						},
						"color": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "A 6 character hex code, without the leading '#', identifying the color of the label.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A short description of the label.",
						},
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL to the issue label.",
						},
					},
				},
			},
		},
	}
}

func resourceGithubIssueLabelsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repository := d.Id()
	ctx := context.WithValue(context.Background(), ctxId, repository)

	log.Printf("[DEBUG] Reading GitHub issue labels for %s/%s", owner, repository)

	labels, err := listLabels(client, ctx, owner, repository)
	if err != nil {
		return err
	}

	err = d.Set("repository", repository)
	if err != nil {
		return err
	}

	err = d.Set("label", flattenLabels(labels))
	if err != nil {
		return err
	}

	return nil
}

func resourceGithubIssueLabelsCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repository := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, repository)

	wantLabels := d.Get("label").(*schema.Set).List()

	wantLabelsMap := make(map[string]interface{}, len(wantLabels))
	for _, label := range wantLabels {
		name := label.(map[string]interface{})["name"].(string)
		if _, found := wantLabelsMap[name]; found {
			return fmt.Errorf("duplicate set label: %s", name)
		}
		wantLabelsMap[name] = label
	}

	hasLabels, err := listLabels(client, ctx, owner, repository)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Updating GitHub issue labels for %s/%s", owner, repository)

	hasLabelsMap := make(map[string]struct{}, len(hasLabels))
	for _, hasLabel := range hasLabels {
		name := hasLabel.GetName()
		wantLabel, found := wantLabelsMap[name]
		if found {
			labelData := wantLabel.(map[string]interface{})
			description := labelData["description"].(string)
			color := labelData["color"].(string)
			if hasLabel.GetDescription() != description || hasLabel.GetColor() != color {
				log.Printf("[DEBUG] Updating GitHub issue label %s/%s/%s", owner, repository, name)

				_, _, err := client.Issues.EditLabel(ctx, owner, repository, name, &github.Label{
					Name:        github.String(name),
					Description: github.String(description),
					Color:       github.String(color),
				})
				if err != nil {
					return err
				}
			}
		} else {
			log.Printf("[DEBUG] Deleting GitHub issue label %s/%s/%s", owner, repository, name)

			_, err := client.Issues.DeleteLabel(ctx, owner, repository, name)
			if err != nil {
				return err
			}
		}

		hasLabelsMap[name] = struct{}{}
	}

	for _, l := range wantLabels {
		labelData := l.(map[string]interface{})
		name := labelData["name"].(string)

		_, found := hasLabelsMap[name]
		if !found {
			log.Printf("[DEBUG] Creating GitHub issue label %s/%s/%s", owner, repository, name)

			_, _, err := client.Issues.CreateLabel(ctx, owner, repository, &github.Label{
				Name:        github.String(name),
				Description: github.String(labelData["description"].(string)),
				Color:       github.String(labelData["color"].(string)),
			})
			if err != nil {
				return err
			}
		}
	}

	d.SetId(repository)

	err = d.Set("label", wantLabels)
	if err != nil {
		return err
	}

	return nil
}

func resourceGithubIssueLabelsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repository := d.Get("repository").(string)
	ctx := context.WithValue(context.Background(), ctxId, repository)

	labels := d.Get("label").(*schema.Set).List()

	log.Printf("[DEBUG] Deleting GitHub issue labels for %s/%s", owner, repository)

	// delete
	for _, raw := range labels {
		label := raw.(map[string]interface{})
		name := label["name"].(string)

		log.Printf("[DEBUG] Deleting GitHub issue label %s/%s/%s", owner, repository, name)

		_, err := client.Issues.DeleteLabel(ctx, owner, repository, name)
		if err != nil {
			return err
		}
	}

	d.SetId(repository)

	err := d.Set("label", make([]map[string]interface{}, 0))
	if err != nil {
		return err
	}

	return nil
}
